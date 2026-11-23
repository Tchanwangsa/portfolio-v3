const SHORT_MONTHS = [
  "Jan",
  "Feb",
  "Mar",
  "Apr",
  "May",
  "Jun",
  "Jul",
  "Aug",
  "Sep",
  "Oct",
  "Nov",
  "Dec",
] as const;

const LONG_MONTHS = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December",
] as const;

function parseIsoDate(isoDate: string): { year: number; month: number; day: number } {
  const [y, m, d] = isoDate.split("-").map(Number);
  return {
    year: y || 0,
    month: m || 1,
    day: d || 1,
  };
}

export function getYear(isoDate: string): string {
  return String(parseIsoDate(isoDate).year);
}

export function formatShortDate(isoDate: string): string {
  const { month, day } = parseIsoDate(isoDate);
  const label = SHORT_MONTHS[month - 1] ?? "";
  return `${label} ${day}`;
}

export function formatFullDate(isoDate: string): string {
  const { year, month, day } = parseIsoDate(isoDate);
  const label = LONG_MONTHS[month - 1] ?? "";
  return `${label} ${day}, ${year}`;
}

export function compareDatesDesc(a: string, b: string): number {
  return b.localeCompare(a);
}
