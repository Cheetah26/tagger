<script lang="ts">
  import type { Tag } from "$bindings/pkg/tagger";
  import EditableTag from "./EditableTag.svelte";
  import { appState } from "./state.svelte";
  import TagTree from "./TagTree.svelte";

  let {
    tag,
    parentMustBeOpen = $bindable(true),
  }: {
    tag: Tag;
    parentMustBeOpen?: boolean;
  } = $props();

  let selected = $derived(appState.selectedTags.some((t) => t.id == tag.id));

  let opened = $state(false);
  let mustBeOpen = $state(false);

  let childTags = $derived(
    appState.tagIdsOrdered
      .filter((id) => tag.children.includes(id))
      .map((childId) => appState.allTags[childId]),
  );

  $effect(() => {
    parentMustBeOpen = selected || mustBeOpen;
  });

  async function toggleSelected() {
    if (selected) {
      appState.selectedTags = appState.selectedTags.filter(
        (t) => t.id != tag.id,
      );
    } else {
      appState.selectedTags.push(tag);
    }
    appState.getFiles();
  }

  function toggleOpened() {
    opened = mustBeOpen || !opened;
  }
</script>

<li>
  <span class="flex flex-row items-center justify-start text-nowrap">
    {#if childTags.length > 0}
      <button onclick={toggleOpened} class="grow text-left">
        <svg
          viewBox="0 0 100 100"
          class="h-3 w-3 inline duration-300 m-1 p-0 fill-none stroke-black stroke-[1rem] {opened
            ? 'rotate-180'
            : ''}"
          focusable="false"
        >
          <polygon points="50 15, 100 100, 0 100" />
        </svg>
        <EditableTag {tag}>{tag.name} ({tag.children.length})</EditableTag>
      </button>
    {:else}
      <EditableTag {tag} class="grow">{tag.name}</EditableTag>
    {/if}

    <input
      type="checkbox"
      id={String(tag.id)}
      checked={selected}
      onchange={toggleSelected}
      class="m-1"
    />
  </span>
  {#if opened}
    <ul class="pl-4">
      {#each childTags as child}
        <TagTree tag={child} bind:parentMustBeOpen={mustBeOpen}></TagTree>
      {/each}
    </ul>
  {/if}
</li>
