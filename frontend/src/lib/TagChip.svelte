<script lang="ts">
  import { getTagString } from "./lib";
  import store from "./store";
  import type { main } from "./wailsjs/go/models";

  export let tag: main.Tag;

  export let clickAction: (() => void) | undefined = undefined;
  export let cancelAction: (() => void) | undefined = undefined;

  $: selected = $store.currentTags.includes(tag);
</script>

<button
  on:click={clickAction}
  class={`relative m-1 p-1 border-2 border-black text-sm text-left ${selected ? "bg-green-300" : ""} ${cancelAction != undefined ? "pr-4" : ""}`}
>
  <p>{tag.name}</p>
  {#if tag.parents}
    <div class="text-xs">
      {#each tag.parents as parent}
        <span>
          {getTagString(parent)}
        </span>
      {/each}
    </div>
  {/if}

  {#if cancelAction != undefined}
    <button on:click={cancelAction} class="absolute top-0 right-0 w-4 h-4"
      >X</button
    >
  {/if}
</button>
