// CREDITS: https://github.com/gpxl-dev/truncate-eth-address/blob/14351c2cd4b342a7fc967f3f389c40cd28f0e94c/src/index.ts
// Captures 0x + 4 characters, then the last 4 characters.
const truncateRegex = /^(0x[a-zA-Z0-9]{4})[a-zA-Z0-9]+([a-zA-Z0-9]{4})$/;

/**
 * Truncates an ethereum address to the format 0x0000…0000
 * @param address Full address to truncate
 * @returns Truncated address
 */
const truncateEthAddress = (address: string) => {
  const match = address.match(truncateRegex);
  if (!match) return address;
  return `${match[1]}...${match[2]}`;
};

export default truncateEthAddress;
