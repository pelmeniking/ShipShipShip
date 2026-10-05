<script lang="ts">
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import { authStore } from "$lib/stores/auth";
    import type { User, UserRole } from "$lib/types";
    import { Loader2, Plus, Trash2, KeyRound, Users } from "lucide-svelte";
    import { Button, Input, Badge, Switch } from "$lib/components/ui";
    import { toast } from "svelte-sonner";
    import * as m from "$lib/paraglide/messages";

    let loading = true;
    let users: User[] = [];

    // New user form
    let showCreateForm = false;
    let creating = false;
    let newUsername = "";
    let newDisplayName = "";
    let newEmail = "";
    let newPassword = "";
    let newRole: UserRole = "editor";

    // Password reset
    let resetUserId: number | null = null;
    let resetPassword = "";

    $: currentUserId = $authStore.user?.id;

    const selectClass =
        "h-9 text-sm px-3 rounded-md border bg-background hover:bg-accent transition-colors focus:outline-none focus:ring-2 focus:ring-ring cursor-pointer disabled:cursor-not-allowed disabled:opacity-50";

    function roleLabel(role: UserRole) {
        return role === "admin" ? m.users_role_admin() : m.users_role_editor();
    }

    function formatDate(value: string | null) {
        if (!value) return m.users_never();
        return new Date(value).toLocaleString();
    }

    async function loadUsers() {
        try {
            users = await api.getUsers();
        } catch (err) {
            toast.error(err instanceof Error ? err.message : String(err));
        } finally {
            loading = false;
        }
    }

    function resetCreateForm() {
        newUsername = "";
        newDisplayName = "";
        newEmail = "";
        newPassword = "";
        newRole = "editor";
        showCreateForm = false;
    }

    async function createUser() {
        if (!newUsername || newPassword.length < 8) {
            toast.error(m.users_validation_required());
            return;
        }
        creating = true;
        try {
            await api.createUser({
                username: newUsername,
                display_name: newDisplayName,
                email: newEmail,
                password: newPassword,
                role: newRole,
            });
            toast.success(m.users_created());
            resetCreateForm();
            await loadUsers();
        } catch (err) {
            toast.error(err instanceof Error ? err.message : String(err));
        } finally {
            creating = false;
        }
    }

    async function updateUser(
        user: User,
        changes: { role?: UserRole; active?: boolean; password?: string },
    ) {
        try {
            const updated = await api.updateUser(user.id, changes);
            users = users.map((u) => (u.id === updated.id ? updated : u));
            toast.success(m.users_updated());
            return true;
        } catch (err) {
            toast.error(err instanceof Error ? err.message : String(err));
            // Reload to undo optimistic UI changes (e.g. the switch)
            await loadUsers();
            return false;
        }
    }

    async function savePasswordReset(user: User) {
        if (resetPassword.length < 8) {
            toast.error(m.users_password_too_short());
            return;
        }
        if (await updateUser(user, { password: resetPassword })) {
            resetUserId = null;
            resetPassword = "";
        }
    }

    async function deleteUser(user: User) {
        if (!confirm(m.users_delete_confirm({ username: user.username }))) {
            return;
        }
        try {
            await api.deleteUser(user.id);
            users = users.filter((u) => u.id !== user.id);
            toast.success(m.users_deleted());
        } catch (err) {
            toast.error(err instanceof Error ? err.message : String(err));
        }
    }

    onMount(loadUsers);
</script>

<svelte:head>
    <title>{m.users_page_title()}</title>
</svelte:head>

<div class="w-full max-w-5xl">
    <div class="mb-8 flex items-start justify-between gap-4">
        <div>
            <h1 class="text-xl font-semibold mb-1">{m.users_heading()}</h1>
            <p class="text-muted-foreground text-sm">
                {m.users_subheading()}
            </p>
        </div>
        {#if !showCreateForm}
            <Button size="sm" on:click={() => (showCreateForm = true)}>
                <Plus class="h-4 w-4 me-2" />
                {m.users_add()}
            </Button>
        {/if}
    </div>

    {#if showCreateForm}
        <form
            class="mb-8 rounded-lg border p-4 space-y-4"
            on:submit|preventDefault={createUser}
        >
            <h3 class="text-base font-semibold">{m.users_add()}</h3>
            <div class="grid gap-4 sm:grid-cols-2">
                <div>
                    <label for="new-username" class="text-sm font-medium mb-2 block"
                        >{m.users_username()}</label
                    >
                    <Input
                        id="new-username"
                        bind:value={newUsername}
                        autocomplete="off"
                        required
                    />
                </div>
                <div>
                    <label
                        for="new-display-name"
                        class="text-sm font-medium mb-2 block"
                        >{m.users_display_name()}</label
                    >
                    <Input id="new-display-name" bind:value={newDisplayName} />
                </div>
                <div>
                    <label for="new-email" class="text-sm font-medium mb-2 block"
                        >{m.users_email()}</label
                    >
                    <Input id="new-email" type="email" bind:value={newEmail} />
                </div>
                <div>
                    <label for="new-role" class="text-sm font-medium mb-2 block"
                        >{m.users_role()}</label
                    >
                    <select
                        id="new-role"
                        bind:value={newRole}
                        class="{selectClass} w-full h-10"
                    >
                        <option value="editor">{m.users_role_editor()}</option>
                        <option value="admin">{m.users_role_admin()}</option>
                    </select>
                </div>
                <div class="sm:col-span-2">
                    <label for="new-password" class="text-sm font-medium mb-2 block"
                        >{m.users_password()}</label
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
            </div>
            <p class="text-xs text-muted-foreground">
                {m.users_role_description()}
            </p>
            <div class="flex gap-2">
                <Button type="submit" size="sm" disabled={creating}>
                    {#if creating}
                        <Loader2 class="h-4 w-4 me-2 animate-spin" />
                    {/if}
                    {m.users_create()}
                </Button>
                <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    on:click={resetCreateForm}
                >
                    {m.users_cancel()}
                </Button>
            </div>
        </form>
    {/if}

    {#if loading}
        <div class="flex items-center justify-center py-16">
            <div class="flex items-center gap-2 text-sm">
                <Loader2 class="h-4 w-4 animate-spin" />
                <span class="text-muted-foreground">{m.users_loading()}</span>
            </div>
        </div>
    {:else if users.length === 0}
        <div class="text-center py-16 text-muted-foreground">
            <Users class="h-8 w-8 mx-auto mb-2" />
            <p class="text-sm">{m.users_empty()}</p>
        </div>
    {:else}
        <div class="rounded-lg border divide-y">
            {#each users as user (user.id)}
                {@const isSelf = user.id === currentUserId}
                <div class="p-4 space-y-3">
                    <div class="flex flex-wrap items-center gap-4">
                        <div class="flex-1 min-w-[200px]">
                            <div class="flex items-center gap-2">
                                <span class="font-medium text-sm"
                                    >{user.display_name || user.username}</span
                                >
                                {#if isSelf}
                                    <Badge variant="secondary">{m.users_you()}</Badge>
                                {/if}
                                {#if !user.active}
                                    <Badge variant="outline">{m.users_inactive()}</Badge>
                                {/if}
                            </div>
                            <div class="text-xs text-muted-foreground mt-0.5">
                                {user.username}
                                {#if user.email}· {user.email}{/if}
                                · {m.users_last_login()}: {formatDate(
                                    user.last_login_at,
                                )}
                            </div>
                        </div>

                        <select
                            class={selectClass}
                            value={user.role}
                            disabled={isSelf}
                            title={roleLabel(user.role)}
                            on:change={(e) =>
                                updateUser(user, {
                                    role: e.currentTarget.value as UserRole,
                                })}
                        >
                            <option value="editor">{m.users_role_editor()}</option>
                            <option value="admin">{m.users_role_admin()}</option>
                        </select>

                        <div class="flex items-center gap-2">
                            <Switch
                                checked={user.active}
                                disabled={isSelf}
                                on:change={(e) =>
                                    updateUser(user, { active: e.detail })}
                            />
                            <span class="text-xs text-muted-foreground"
                                >{m.users_active()}</span
                            >
                        </div>

                        <Button
                            variant="ghost"
                            size="sm"
                            title={m.users_reset_password()}
                            on:click={() => {
                                resetUserId =
                                    resetUserId === user.id ? null : user.id;
                                resetPassword = "";
                            }}
                        >
                            <KeyRound class="h-4 w-4" />
                        </Button>
                        <Button
                            variant="ghost"
                            size="sm"
                            title={m.users_delete()}
                            disabled={isSelf}
                            on:click={() => deleteUser(user)}
                        >
                            <Trash2 class="h-4 w-4 text-destructive" />
                        </Button>
                    </div>

                    {#if resetUserId === user.id}
                        <form
                            class="flex flex-wrap items-center gap-2"
                            on:submit|preventDefault={() =>
                                savePasswordReset(user)}
                        >
                            <Input
                                type="password"
                                class="max-w-xs h-9"
                                placeholder={m.users_new_password()}
                                autocomplete="new-password"
                                bind:value={resetPassword}
                            />
                            <Button type="submit" size="sm"
                                >{m.users_save_password()}</Button
                            >
                            <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                on:click={() => (resetUserId = null)}
                                >{m.users_cancel()}</Button
                            >
                        </form>
                    {/if}
                </div>
            {/each}
        </div>
    {/if}
</div>
