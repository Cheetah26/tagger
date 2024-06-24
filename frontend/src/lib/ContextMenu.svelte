<script lang="ts">
    // pos is cursor position when right click occur
    let pos = { x: 0, y: 0 };
    // menu is dimension (height and width) of context menu
    let menu = { h: 0, w: 0 };
    // browser/window dimension (height and width)
    let browser = { h: 0, w: 0 };
    // showMenu is state of context-menu visibility
    let showMenu = false;
    // to display some text
    let content;

    function openMenu(e: MouseEvent) {
        window.dispatchEvent(new Event("click"));

        browser = {
            w: window.innerWidth,
            h: window.innerHeight,
        };
        pos = {
            x: e.clientX,
            y: e.clientY,
        };
        // If bottom part of context menu will be displayed
        // after right-click, then change the position of the
        // context menu. This position is controlled by `top` and `left`
        // at inline style.
        // Instead of context menu is displayed from top left of cursor position
        // when right-click occur, it will be displayed from bottom left.
        if (browser.h - pos.y < menu.h) pos.y = pos.y - menu.h;
        if (browser.w - pos.x < menu.w) pos.x = pos.x - menu.w;
        showMenu = true;
    }

    function onPageClick() {
        // To make context menu disappear when
        // mouse is clicked outside context menu
        showMenu = false;
    }

    function getContextMenuDimension(node: HTMLElement) {
        // This function will get context menu dimension
        // when navigation is shown => showMenu = true
        let height = node.offsetHeight;
        let width = node.offsetWidth;
        menu = {
            h: height,
            w: width,
        };
    }

    export let menuItems: {
        name: String;
        onClick: () => void | Promise<void>;
    }[];
</script>

<svelte:window on:click={onPageClick} />

<menu on:contextmenu|preventDefault={openMenu} class="inline">
    <slot></slot>
    {#if showMenu}
        <div
            use:getContextMenuDimension
            style="top:{pos.y}px; left:{pos.x}px"
            class="bg-white absolute border-2 border-black rounded-md p-1"
        >
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
        </div>
    {/if}
</menu>
