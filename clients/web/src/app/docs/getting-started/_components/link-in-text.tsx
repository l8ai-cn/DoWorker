import Link from "next/link";

interface LinkInTextProps {
  raw: string;
  linkHref: string;
  linkLabel: string;
}

export function LinkInText({ raw, linkHref, linkLabel }: LinkInTextProps) {
  const parts = raw.split("{link}");
  if (parts.length < 2) return <>{raw}</>;
  return (
    <>
      {parts[0]}
      <Link href={linkHref} className="text-primary hover:underline">
        {linkLabel}
      </Link>
      {parts[1]}
    </>
  );
}
