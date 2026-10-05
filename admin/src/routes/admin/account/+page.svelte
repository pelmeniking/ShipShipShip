<script lang="ts">
    import { api } from "$lib/api";
    import { authStore } from "$lib/stores/auth";
    import { Loader2, KeyRound } from "lucide-svelte";
    import { Button, Input } from "$lib/components/ui";
    import { toast } from "svelte-sonner";
    import * as m from "$lib/paraglide/messages";

    let currentPassword = "";
    let newPassword = "";
    let confirmPassword = "";
    let saving = false;

    $: user = $authStore.user;

    async function changePassword() {
        if (newPassword.length < 8) {
            toast.error(m.users_password_too_short());
            return;
        }
        if (newPassword !== confirmPassword) {
            toast.error(m.account_password_mismatch());
            return;
        }
        saving = true;
        try {
            await api.changeOwnPassword(currentPassword, newPassword);
            toast.success(m.account_password_changed());
            currentPassword = "";
            newPassword = "";
            confirmPassword = "";
        } catch (err) {
            toast.error(err instanceof Error ? err.message : String(err));
        } finally {
            saving = false;
        }
    }
</script>

<svelte:head>
    <title>{m.account_page_title()}</title>
</svelte:head>

<div class="w-full max-w-xl">
    <div class="mb-8">
        <h1 class="text-xl font-semibold mb-1">{m.account_heading()}</h1>
        {#if user}
            <p class="text-muted-foreground text-sm">
                {user.display_name || user.username} · {user.username}
                {#if user.email}· {user.email}{/if}
                · {user.role === "admin"
                    ? m.users_role_admin()
                    : m.users_role_editor()}
            </p>
        {/if}
    </div>

    {#if $authStore.isDemoMode}
        <p class="text-sm text-muted-foreground">{m.account_demo_mode()}</p>
    {:else}
        <form class="space-y-4" on:submit|preventDefault={changePassword}>
            <div class="flex items-center gap-3 mb-1.5">
                <KeyRound class="h-5 w-5 text-primary" />
                <h3 class="text-base font-semibold">
                    {m.account_change_password()}
                </h3>
            </div>
            <div>
                <label for="current-password" class="text-sm font-medium mb-2 block"
                    >{m.account_current_password()}</label
                >
                <Input
                    id="current-password"
                    type="password"
                    bind:value={currentPassword}
                    autocomplete="current-password"
                    required
                />
            </div>
            <div>
                <label for="new-password" class="text-sm font-medium mb-2 block"
                    >{m.users_new_password()}</label
                >
                <Input
                    id="new-password"
                    type="password"
                    bind:value={newPassword}
                    autocomplete="new-password"
                    required
                />
                <p class="text-xs text-muted-foreground mt-1.5">
                    {m.users_password_hint()}
                </p>
            </div>
            <div>
                <label for="confirm-password" class="text-sm font-medium mb-2 block"
                    >{m.account_confirm_password()}</label
                >
                <Input
                    id="confirm-password"
                    type="password"
                    bind:value={confirmPassword}
                    autocomplete="new-password"
                    required
                />
            </div>
            <Button type="submit" size="sm" disabled={saving}>
                {#if saving}
                    <Loader2 class="h-4 w-4 me-2 animate-spin" />
                {/if}
                {m.users_save_password()}
            </Button>
        </form>
    {/if}
</div>
