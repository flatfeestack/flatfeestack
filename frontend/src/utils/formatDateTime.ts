export default function formatDateTime(date: Date): string {
  return new Intl.DateTimeFormat(navigator.language, {
    weekday: "short",
    year: "numeric",
    month: "numeric",
    day: "numeric",
    hour: "numeric",
    minute: "numeric",
    second: "numeric",
  }).format(date);
}
