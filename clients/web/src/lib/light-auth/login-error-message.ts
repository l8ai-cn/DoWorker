import { ApiError } from "@/lib/api/api-types";

type LoginErrorT = (key: string) => string;

export function resolveLoginErrorMessage(err: unknown, t: LoginErrorT): string {
  if (err instanceof ApiError) {
    if (/sso/i.test(err.serverMessage ?? "")) {
      return t("auth.sso.ssoRequired");
    }
    if (err.status === 401 || err.code === "UNAUTHENTICATED") {
      return t("auth.loginPage.invalidCredentials");
    }
    if (err.status >= 500 || err.status === 502 || err.status === 503) {
      return t("auth.loginPage.serverUnavailable");
    }
  }
  if (err instanceof TypeError) {
    return t("auth.loginPage.serverUnavailable");
  }
  return t("common.error");
}
