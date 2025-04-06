<script lang="ts">
  import type { Tag } from "$bindings/pkg/tagger";
  import ContextMenu from "$lib/components/ContextMenu.svelte";
  import Modal from "$lib/components/Modal.svelte";
  import TagEditor from "$lib/TagEditor.svelte";
  import store from "$lib/store";
  import { get } from "svelte/store";
  import type { Snippet } from "svelte";

  let { tag, children }: { tag: Tag; children: Snippet } = $props();

  let showEditModal = $state(false);
  let editedTag = $state(tag);

  async function submitEdit() {
    await store.updateTag(editedTag);
    showEditModal = false;
  }

  function cancelEdit() {
    showEditModal = false;
    editedTag = tag;
  }

  async function addTagTag(tag: Tag) {
    if (!editedTag.parents) {
      editedTag.parents = [];
    }
    editedTag.parents = [...editedTag.parents, tag.id];
  }

  async function removeTagTag(tag: Tag) {
    console.log(editedTag.parents.filter((id) => id != tag.id));
    editedTag.parents = editedTag.parents.filter((id) => id != tag.id);
  }
</script>

<ContextMenu
  menuItems={[
    {
      name: "Edit Tag",
      onClick: () => {
        showEditModal = true;
      },
    },
    {
      name: "Delete Tag",
      onClick: () => {
        if (confirm("Really delete tag " + tag.name + "?")) {
          store.removeTag(tag);
        }
      },
    },
  ]}>{@render children()}</ContextMenu
>

<Modal bind:open={showEditModal}>
  <h1>Edit tag</h1>
  <form onsubmitcapture={submitEdit}>
    <p>Id: {editedTag.id}</p>
    <label for="tag-name">Name:</label>
    <input type="text" id="tag-name" bind:value={editedTag.name} />

    <TagEditor
      tags={editedTag.parents.map((t) => get(store).tags[t])}
      onAdd={addTagTag}
      onRemove={removeTagTag}
    ></TagEditor>

    <button onclick={cancelEdit}>Cancel</button>
    <button type="submit">Submit</button>
  </form>
</Modal>
