import { derived, writable } from "svelte/store";
import { api } from "$lib/api";
import type { CurrentUser } from "$lib/types";

let isDemoMode = false;

export interface AuthState {
  isAuthenticated: boolean;
  loading: boolean;
  initialized: boolean;
  isDemoMode: boolean;
  user?: CurrentUser;
}

// Create the auth store with initial state
function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>({
    isAuthenticated: false,
    loading: true,
    initialized: false,
    isDemoMode: false,
  });

  return {
    subscribe,

    // Initialize authentication state
    async init() {
      update((state) => ({ ...state, loading: true }));

      try {
        // Check if demo mode is enabled
        const demoResponse = await api.checkDemoMode();
        isDemoMode = demoResponse.demo_mode;

        if (isDemoMode) {
          // In demo mode, bypass authentication
          set({
            isAuthenticated: true,
            loading: false,
            initialized: true,
            isDemoMode: true,
            user: {
              username: "demo",
              display_name: "Demo User",
              role: "admin",
            },
          });
          return true;
        }

        if (!api.isAuthenticated()) {
          set({
            isAuthenticated: false,
            loading: false,
            initialized: true,
            isDemoMode: false,
          });
          return false;
        }

        const result = await api.validateToken();
        set({
          isAuthenticated: true,
          loading: false,
          initialized: true,
          isDemoMode: false,
          user: {
            id: result.id,
            username: result.username,
            display_name: result.display_name,
            email: result.email,
            role: result.role,
          },
        });
        return true;
      } catch {
        api.clearToken();
        set({
          isAuthenticated: false,
          loading: false,
          initialized: true,
          isDemoMode: false,
        });
        return false;
      }
    },

    // Set authenticated state after login
    setAuthenticated(user?: CurrentUser) {
      set({
        isAuthenticated: true,
        loading: false,
        initialized: true,
        isDemoMode: false,
        user,
      });
    },

    // Clear authentication state
    logout() {
      if (!isDemoMode) {
        api.clearToken();
      }
      set({
        isAuthenticated: false,
        loading: false,
        initialized: true,
        isDemoMode: false,
      });
    },

    // Reset to initial state
    reset() {
      set({
        isAuthenticated: false,
        loading: true,
        initialized: false,
        isDemoMode: false,
      });
    },
  };
}

export const authStore = createAuthStore();

// True when the signed-in user may manage settings, appearance, newsletter and users
export const isAdmin = derived(
  authStore,
  ($auth) => $auth.isDemoMode || $auth.user?.role === "admin",
);
