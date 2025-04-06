<script lang="ts">
  import type { Tag } from "$bindings/pkg/tagger";
  import EditableTag from "./EditableTag.svelte";
  import store from "./store";
  import TagTree from "./TagTree.svelte";

  let {
    tag,
    parentMustBeOpen = $bindable(true),
  }: {
    tag: Tag;
    parentMustBeOpen?: boolean;
  } = $props();

  let selected = $derived($store.currentTags.includes(tag));
  function toggleSelected() {
    if (selected) {
      store.deselectTag(tag);
    } else {
      store.selectTag(tag);
    }
  }

  $effect(() => {
    parentMustBeOpen = selected || mustBeOpen;
  });

  let mustBeOpen = $state(false);

  let opened = $state(false);
  function toggleOpened() {
    opened = mustBeOpen || !opened;
  }

  let childTags = $derived(tag.children.map((c) => $store.tags[c]));
</script>

<li>
  <p class="flex flex-row items-center text-nowrap">
    {#if childTags.length > 0}
      <svg
        id="triangle"
        viewBox="0 0 100 100"
        class="w-2 h-2 {opened ? 'rotate-180' : ''} duration-200"
      >
        <polygon points="50 15, 100 100, 0 100" />
      </svg>

      <EditableTag {tag}>
        <button onclick={toggleOpened}
          ><span class={selected ? "bg-green-500" : "bg-white"}
            >{tag.name} ({tag.children.length})</span
          ></button
        ></EditableTag
      >
    {:else}
      <EditableTag {tag}
        ><span class={selected ? "bg-green-500" : "bg-white"}>{tag.name}</span
        ></EditableTag
      >
    {/if}
    <span class="flex-grow"></span>
    <button onclick={toggleSelected} class="px-1 border">+</button>
  </p>
  {#if opened}
    <ul class="pl-4">
      {#each childTags as child}
        <TagTree tag={child} bind:parentMustBeOpen={mustBeOpen}></TagTree>
      {/each}
    </ul>
  {/if}
</li>
