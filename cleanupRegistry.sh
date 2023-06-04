#!/bin/bash

# Check if jq is installed
if ! command -v jq >/dev/null 2>&1; then
  echo "jq is not installed. Please install jq to run this script."
  exit 1
fi

DIGITALOCEAN_TOKEN="$DIGITALOCEAN_TOKEN"
URL1="https://api.digitalocean.com/v2/registry/flatfeestack/garbage-collection"
URL2="https://api.digitalocean.com/v2/registry/flatfeestack/garbage-collections"
URL3="https://api.digitalocean.com/v2/registry/flatfeestack/repositoriesV2?page_size=1"

start_garbage_collection() {
  # Execute the POST request to start garbage collection
  response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" \
    -w "%{http_code}" \
    "$URL1")

  # Extract the HTTP response code from the response
  response_code="${response:${#response}-3}"

  # Check the HTTP response code
  if [[ $response_code != 200 && $response_code != 201 ]]; then
    echo "Abort! Error: $response"
    return 1
  fi

  # Remove the last line (response code) from the response
  response_body=$(echo "$response" | sed '$d')

  # Extract the UUID from the response
  uuid=$(echo "$response_body" | jq -r '.garbage_collection.uuid')
  echo "Garbage collection started with UUID: $uuid"
}

check_status() {
  local response_code=""
  local status=""
  local updated_at=""

  for ((i = 0; i < 30; i++)); do
    # Execute the GET request to check garbage collection status
    response=$(curl -s -X GET \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" \
      -w "\n%{http_code}" \
      "$URL2")

    # Extract the HTTP response code from the response
    response_code=$(echo "$response" | tail -n 1)

    # Check the HTTP response code
    if [[ $response_code != 200 && $response_code != 201 ]]; then
      echo "Abort! Error: $response"
      return 1
    fi

    # Remove the last line (response code) from the response
    response_body=$(echo "$response" | sed '$d')

    # Check if the UUID is present in the garbage_collections list
    status=$(echo "$response_body" | jq -r '.garbage_collections[] | select(.uuid == "'"$uuid"'") | .status')

    # Check if the garbage collection has finished
    if [[ "$status" == "succeeded" ]]; then
      echo "Garbage collection with UUID $uuid succeeded!"
      return 0
    elif [[ "$status" == "cancelling" || "$status" == "failed" || "$status" == "cancelled" ]]; then
      echo "Garbage collection with UUID $uuid aborted! Status: $status"
      return 1
    fi

    echo "Waiting..."
    sleep 10
  done

  echo "Timeout! Garbage collection with UUID $uuid not finished."
  return 1
}

delete_digest() {
  local name="$1"
  local digest="$2"

  # Execute the DELETE request to delete the digest
  response=$(curl -s -X DELETE \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" \
    -w "%{http_code}" \
    "https://api.digitalocean.com/v2/registry/flatfeestack/repositories/$name/digests/$digest")

  # Extract the HTTP response code from the response
  response_code="${response:${#response}-3}"

  # Check the HTTP response code
  if [[ $response_code != 204 ]]; then
    echo "Failed to delete digest $digest for repository $name. Error: $response"
  else
    echo "Deleted digest $digest for repository $name."
  fi
}

is_date_older_than_10_days() {
  local date_string=$1
  local ten_days_ago=$(date -v -10d -Iseconds)

  if [[ $date_string < $ten_days_ago ]]; then
    return 0
  else
    return 1
  fi
}

delete_old_tag() {
  local name="$1"

  response=$(curl -s -X GET \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" \
    -w "\n%{http_code}" \
    "https://api.digitalocean.com/v2/registry/flatfeestack/repositories/$name/tags")

  # Extract the HTTP response code from the response
  response_code=$(echo "$response" | tail -n 1)

  # Check the HTTP response code
  if [[ $response_code != 200 && $response_code != 201 ]]; then
    echo "Abort! Error: $response"
    return 1
  fi

  # Remove the last line (response code) from the response
  response_body=$(echo "$response" | sed '$d')

  # Extract the tags from the response using jq
  tags=$(echo "$response_body" | jq -r '.tags[] | select(.tag != null and .tag != "" and .tag != "main") | .tag')

  for tag in $tags; do
    # Extract the updated_at value for the tag using jq
    echo "Checking tag: $tag"
    updated_at=$(echo "$response_body" | jq -r --arg tag "$tag" '.tags[] | select(.tag == $tag) | .updated_at')

    # Check if the updated_at is older than 10 days
    if is_date_older_than_10_days "$updated_at"; then
      echo "Deleting tag: $tag"

      # Make the API call to delete the tag
      delete_response=$(curl -s -X DELETE \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" \
        -w "\n%{http_code}" \
        "https://api.digitalocean.com/v2/registry/flatfeestack/repositories/$name/tags/$tag")

      # Extract the HTTP response code from the response
      response_code=$(echo "$delete_response" | tail -n 1)

      # Check the HTTP response code
      if [[ $response_code != 204 ]]; then
        echo "Abort! Error: $response"
        return 1
      fi
    fi
  done

}

find_and_delete() {
  repositories_response=$(curl -s -X GET \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" \
    "$URL3")

  # Extract the names from the repositories list
  names=$(echo "$repositories_response" | jq -r '.repositories[].name')

  # Loop over the names
  for name in $names; do
    echo "Name: $name"

    delete_old_tag "$name"

    # Start with the first page
    page=1
    has_next_page=true

    # Loop through the pages
    while [ "$has_next_page" = true ]; do
      digests_response=$(curl -s -X GET \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" \
        "https://api.digitalocean.com/v2/registry/flatfeestack/repositories/$name/digests?page=$page&per_page=20")

      # Extract the digests from the manifests list
      empty_digests=$(echo "$digests_response" | jq -r '.manifests[] | select(.tags == []) | .digest')

      # Loop over the digests
      for digest in $empty_digests; do
        echo "Digest: $digest"
        delete_digest "$name" "$digest"
      done

      # Check if there is a next page
      next_page_url=$(echo "$digests_response" | jq -r '.links.pages.next')
      if [ "$next_page_url" != "null" ]; then
        # Extract the page number from the next page URL
        page=$(echo "$next_page_url" | awk -F'page=' '{print $2}' | awk -F'&' '{print $1}')
      else
        # No more pages available, exit the loop
        has_next_page=false
      fi
    done
  done
}

find_and_delete
start_garbage_collection
check_status
