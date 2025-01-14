<script lang="ts">
    /**
     * List of items to show on the context menu with an onClick callback
     */
    export let menuItems: {
        name: String;
        onClick: () => void | Promise<void>;
    }[];

    /**
     * A parent element used to constrain the position of the context menu
     * If undefined, the window is used
     */
    export let boundingElement: Element | undefined = undefined;

    let pos = { x: 0, y: 0 };
    let showMenu = false;

    function openMenu(mouse: MouseEvent) {
        // close any other menus first
        window.dispatchEvent(new Event("click"));

        pos = {
            x: mouse.x,
            y: mouse.y,
        };

        showMenu = true;
    }

    function positionContextMenu(node: HTMLElement) {
        // size of the context menu
        const height = node.offsetHeight;
        const width = node.offsetWidth;

        // find the bounds
        const maxRight =
            boundingElement?.getBoundingClientRect().right || window.innerWidth;
        const maxBottom =
            boundingElement?.getBoundingClientRect().bottom ||
            window.innerHeight;

        // adjust position if necessary
        if (pos.x + width > maxRight) pos.x = pos.x - width;
        if (pos.y + height > maxBottom) pos.y = pos.y - height;

        // set the position
        node.style.top = pos.y + "px";
        node.style.left = pos.x + "px";
    }
</script>

<!-- Close the menu when the page is clicked -->
<svelte:window on:click={() => (showMenu = false)} />

<menu on:contextmenu|preventDefault={openMenu} class={$$props.class}>
    <slot></slot>
    {#if showMenu}
        <div
            use:positionContextMenu
            class="bg-white z-10 absolute border-2 border-black rounded-md p-1"
        >
            <slot name="menuItems" />
            {#if menuItems}
                <ul>
                    {#each menuItems as item}
                        {#if item.name == "hr"}
                            <hr />
                        {:else}
                            <li>
                                <button on:click={item.onClick} class="text-xs"
                                    >{item.name}</button
                                >
                            </li>
                        {/if}
                    {/each}
                </ul>
            {/if}
        </div>
    {/if}
</menu>
