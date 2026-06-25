"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { AuthStatusShell } from "@/components/auth/AuthStatusShell";

export function LoadingScreen({ title, subtitle }: { title: string; subtitle: string }) {
  return <AuthStatusShell title={title} subtitle={subtitle} variant="loading" />;
}

export function ErrorScreen({
  title,
  error,
  loginLabel,
}: {
  title: string;
  error: string;
  loginLabel: string;
}) {
  return (
    <AuthStatusShell title={title} subtitle={error} variant="error">
      <Link href="/login">
        <Button className="w-full">{loginLabel}</Button>
      </Link>
    </AuthStatusShell>
  );
}

export function ExpiredScreen({
  title,
  description,
  hint,
}: {
  title: string;
  description: string;
  hint: string;
}) {
  return (
    <AuthStatusShell title={title} subtitle={description} variant="none" footer={hint} />
  );
}

export function SuccessScreen({
  title,
  description,
  hint,
}: {
  title: string;
  description: string;
  hint: string;
}) {
  return (
    <AuthStatusShell title={title} subtitle={description} variant="success" footer={hint} />
  );
}
