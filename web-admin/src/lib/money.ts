export function formatINR(paise: number): string {
  const inr = paise / 100;
  return inr.toLocaleString("en-IN", {
    style: "currency",
    currency: "INR",
    maximumFractionDigits: 2
  });
}

