// Dark mode removed — app is light-mode only.
// This stub exists so existing imports don't break.
export type Theme = 'light';

export function useTheme() {
  return { theme: 'light' as Theme, toggle: () => {} };
}
