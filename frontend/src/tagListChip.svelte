<script lang="ts">
  import ContextMenu from "./lib/ContextMenu.svelte";
  import store from "./lib/store";
  import type { main } from "./lib/wailsjs/go/models";

  export let tag: main.Tag;

  const contextMenuItems = [
    {
      name: "Delete Tag",
      onClick: () => {
        if (confirm("Really delete tag " + tag.name + "?")) {
          store.removeTag(tag);
        }
      },
    },
  ];

  const BaseStyle = "border-2 border-black m-1 p-1 text-sm ";
</script>

<ContextMenu menuItems={contextMenuItems}>
  {#if $store.currentTags.includes(tag)}
    <button
      on:click={() => store.deselectTag(tag)}
      class={BaseStyle + "bg-green-300"}>{tag.name}</button
    >
  {:else}
    <button on:click={() => store.selectTag(tag)} class={BaseStyle}
      >{tag.name}</button
    >
  {/if}
</ContextMenu>
