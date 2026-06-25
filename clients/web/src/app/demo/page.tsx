"use client";

import { useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { AuthShell } from "@/components/auth/AuthShell";
import { AuthStatusShell } from "@/components/auth/AuthStatusShell";
import { useTranslations } from "next-intl";

export default function DemoRequestPage() {
  const t = useTranslations();
  const [formData, setFormData] = useState({
    name: "",
    email: "",
    company: "",
    referral: "",
    message: "",
  });
  const [loading, setLoading] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [error, setError] = useState("");

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    setFormData((prev) => ({ ...prev, [e.target.name]: e.target.value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      const subject = encodeURIComponent(`Demo Request from ${formData.name} (${formData.company})`);
      const body = encodeURIComponent(
        `Name: ${formData.name}\nEmail: ${formData.email}\nCompany: ${formData.company}\nHow did you hear about us: ${formData.referral}\n\nMessage:\n${formData.message}`,
      );
      window.location.href = `mailto:bd@agentsmesh.ai?subject=${subject}&body=${body}`;
      setSubmitted(true);
    } catch {
      setError(t("landing.demo.error"));
    } finally {
      setLoading(false);
    }
  };

  if (submitted) {
    return (
      <AuthStatusShell
        title={t("landing.demo.thankYou")}
        subtitle={t("landing.demo.thankYouDescription")}
        variant="success"
        footer={
          <Link href="/" className="text-sm text-[var(--azure-text-muted)] hover:text-foreground motion-interactive">
            {t("landing.demo.backToHome")}
          </Link>
        }
      />
    );
  }

  return (
    <AuthShell title={t("landing.demo.title")} subtitle={t("landing.demo.subtitle")}>
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">{error}</div>
        )}

        <div className="space-y-2">
          <label htmlFor="name" className="text-sm font-medium text-foreground">
            {t("landing.demo.nameLabel")}
          </label>
          <Input
            id="name"
            name="name"
            placeholder={t("landing.demo.namePlaceholder")}
            value={formData.name}
            onChange={handleChange}
            required
          />
        </div>

        <div className="space-y-2">
          <label htmlFor="email" className="text-sm font-medium text-foreground">
            {t("landing.demo.emailLabel")}
          </label>
          <Input
            id="email"
            name="email"
            type="email"
            placeholder={t("landing.demo.emailPlaceholder")}
            value={formData.email}
            onChange={handleChange}
            required
          />
        </div>

        <div className="space-y-2">
          <label htmlFor="company" className="text-sm font-medium text-foreground">
            {t("landing.demo.companyLabel")}
          </label>
          <Input
            id="company"
            name="company"
            placeholder={t("landing.demo.companyPlaceholder")}
            value={formData.company}
            onChange={handleChange}
            required
          />
        </div>

        <div className="space-y-2">
          <label htmlFor="referral" className="text-sm font-medium text-foreground">
            {t("landing.demo.referralLabel")}
          </label>
          <Input
            id="referral"
            name="referral"
            placeholder={t("landing.demo.referralPlaceholder")}
            value={formData.referral}
            onChange={handleChange}
          />
        </div>

        <div className="space-y-2">
          <label htmlFor="message" className="text-sm font-medium text-foreground">
            {t("landing.demo.messageLabel")}
          </label>
          <Textarea
            id="message"
            name="message"
            placeholder={t("landing.demo.messagePlaceholder")}
            value={formData.message}
            onChange={handleChange}
            rows={4}
            className="resize-none"
          />
        </div>

        <Button type="submit" className="w-full azure-gradient-bg text-white border-0" disabled={loading}>
          {loading ? t("landing.demo.submitting") : t("landing.demo.submit")}
        </Button>
      </form>
    </AuthShell>
  );
}
