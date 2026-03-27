import { useEffect, useMemo } from "react";

type Theme = "dark" | "light" | "system";

type ThemeProviderProps = {
  children: React.ReactNode;
  defaultTheme?: Theme;
  storageKey?: string;
};

export function ThemeProvider({
  children,
  defaultTheme = "system",
  storageKey = "diagura-admin-theme",
}: ThemeProviderProps) {
  const theme = useMemo<Theme>(
    () => (localStorage.getItem(storageKey) as Theme) || defaultTheme,
    [storageKey, defaultTheme],
  );

  useEffect(() => {
    const root = window.document.documentElement;
    const media = window.matchMedia("(prefers-color-scheme: dark)");

    function apply() {
      root.classList.remove("light", "dark");
      if (theme === "system") {
        root.classList.add(media.matches ? "dark" : "light");
      } else {
        root.classList.add(theme);
      }
    }

    apply();

    if (theme === "system") {
      media.addEventListener("change", apply);
      return () => media.removeEventListener("change", apply);
    }
  }, [theme]);

  return <>{children}</>;
}
