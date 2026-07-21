"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Loader2 } from "lucide-react";
import { AgentCloudMark } from "@/components/brand/AgentCloudMark";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useAuthStore } from "@/stores/auth";
import { login } from "@/lib/api/admin";
import { toast } from "sonner";

export default function LoginPage() {
  const router = useRouter();
  const { token, setAuth } = useAuthStore();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (token) {
      router.replace("/");
    }
  }, [token, router]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!username || !password) {
      toast.error("请输入用户名和密码");
      return;
    }

    setIsLoading(true);

    try {
      const result = await login({ email: username, password });
      setAuth(result.token, result.refresh_token, result.user);
      toast.success(`欢迎回来，${result.user.name || result.user.username}`);
      router.replace("/");
    } catch (err: unknown) {
      const error = err as { error?: string };
      toast.error(error.error || "登录失败，请检查账号和密码。");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <main className="admin-auth-theme relative flex min-h-screen items-center justify-center overflow-hidden bg-background px-4 py-10 text-foreground">
      <div
        className="pointer-events-none absolute inset-0 opacity-70"
        style={{
          backgroundImage:
            "linear-gradient(to right, rgba(17,24,39,0.07) 1px, transparent 1px), linear-gradient(to bottom, rgba(17,24,39,0.07) 1px, transparent 1px)",
          backgroundSize: "96px 96px",
        }}
      />
      <section className="relative z-10 w-full max-w-md">
        <div className="mb-7 flex items-center justify-center">
          <div className="flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center overflow-hidden rounded-lg shadow-sm">
              <AgentCloudMark className="h-full w-full" />
            </div>
            <span className="text-2xl font-semibold text-foreground">Agent Cloud</span>
          </div>
        </div>

        <Card className="border-border/80 bg-card/95 p-7 shadow-[0_22px_70px_rgba(17,24,39,0.13)] backdrop-blur sm:p-9">
          <CardHeader className="space-y-2 p-0 text-center">
            <CardTitle className="text-3xl font-semibold tracking-normal">管理控制台</CardTitle>
            <CardDescription>
              使用管理员账号登录以访问管理后台。
            </CardDescription>
          </CardHeader>
          <CardContent className="p-0 pt-9">
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <label htmlFor="username" className="text-sm font-medium">
                  用户名
                </label>
                <Input
                  id="username"
                  type="text"
                  placeholder="admin"
                  className="h-10 bg-white"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  disabled={isLoading}
                  autoComplete="username"
                />
              </div>
              <div className="space-y-2">
                <label htmlFor="password" className="text-sm font-medium">
                  密码
                </label>
                <Input
                  id="password"
                  type="password"
                  placeholder="••••••••"
                  className="h-10 bg-white"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  disabled={isLoading}
                  autoComplete="current-password"
                />
              </div>
              <Button
                type="submit"
                className="h-10 w-full shadow-[0_8px_20px_color-mix(in_srgb,var(--primary)_22%,transparent)]"
                size="lg"
                disabled={isLoading}
              >
                {isLoading ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    登录中...
                  </>
                ) : (
                  "登录"
                )}
              </Button>
            </form>
            <p className="mt-4 text-center text-xs text-muted-foreground">
              仅系统管理员可以访问此控制台。
            </p>
          </CardContent>
        </Card>
      </section>
    </main>
  );
}
