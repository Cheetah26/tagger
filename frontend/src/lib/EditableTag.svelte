<script lang="ts">
  import { TaggerService } from "$bindings/index";
  import type { Tag } from "$bindings/pkg/tagger";
  import ContextMenu from "$lib/components/ContextMenu.svelte";
  import Modal from "$lib/components/Modal.svelte";
  import TagEditor from "$lib/TagEditor.svelte";
  import type { Snippet } from "svelte";
  import { appState } from "./state.svelte";

  let {
    tag,
    children,
    boundingElement,
    class: suppliedClass,
  }: {
    tag: Tag;
    children: Snippet;
    boundingElement?: HTMLElement;
    class?: string;
  } = $props();

  let showEditModal = $state(false);

  let newName = $state(tag.name);
  let newChildren = $state(tag.children);
  let newParents = $state(tag.parents);

  async function edit() {
    await TaggerService.UpdateTag({
      id: tag.id,
      name: newName,
      parents: newParents,
      children: newChildren,
    });
    appState.allTags = await TaggerService.GetAllTags();
    showEditModal = false;
  }

  function cancel() {
    newName = tag.name;
    newChildren = tag.children;
    newParents = tag.parents;

    showEditModal = false;
  }

  function addTagTag(tag: Tag) {
    newParents.push(tag.id);
  }

  function removeTagTag(tag: Tag) {
    newParents = newParents.filter((id) => id != tag.id);
  }

  const menuItems = [
    {
      name: "Edit Tag",
      onClick: () => {
        showEditModal = true;
      },
    },
    {
      name: "Delete Tag",
      onClick: async () => {
        if (confirm("Really delete tag " + tag.name + "?")) {
          await TaggerService.RemoveTag(tag);
          appState.allTags = await TaggerService.GetAllTags();

          appState.selectedTags = appState.selectedTags.filter(
            (t) => t.id != tag.id,
          );
          appState.getFiles();
        }
      },
    },
  ];
</script>

<ContextMenu {menuItems} {boundingElement} class={suppliedClass}
  >{@render children()}</ContextMenu
>

<Modal bind:open={showEditModal}>
  <h1>Edit tag (id: {tag.id})</h1>

  <label for="tag-name">Name:</label>
  <input type="text" id="tag-name" bind:value={newName} />

  <TagEditor
    tags={newParents.map((t) => appState.allTags[t])}
    onAdd={addTagTag}
    onRemove={removeTagTag}
  ></TagEditor>

  <button onclick={cancel}>Cancel</button>
  <button onclick={edit}>Submit</button>
</Modal>
