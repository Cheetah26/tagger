<script lang="ts">
  import ContextMenu from "./lib/ContextMenu.svelte";
  import TagEditor from "./lib/TagEditor.svelte";
  import Modal from "./lib/Modal.svelte";
  import store from "./lib/store";
  import type { tagger } from "./lib/wailsjs/go/models";
  import TagChip from "./lib/TagChip.svelte";

  export let contextMenuBounds;

  export let tag: tagger.Tag;

  let editModalOpen = false;
  let editedTag = tag;

  async function submitEdit() {
    await store.updateTag(editedTag);
    editModalOpen = false;
  }

  function cancelEdit() {
    editModalOpen = false;
    editedTag = tag;
  }

  const contextMenuItems = [
    {
      name: "Edit Tag",
      onClick: () => {
        editModalOpen = true;
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
  ];

  async function addTagTag(tag: tagger.Tag) {
    if (!editedTag.parents) {
      editedTag.parents = [];
    }
    editedTag.parents = [...editedTag.parents, tag];
  }

  async function removeTagTag(tag: tagger.Tag) {
    console.log(editedTag.parents.filter((t) => t.id != tag.id));
    editedTag.parents = editedTag.parents.filter((t) => t.id != tag.id);
  }

  function toggleTagSelection() {
    if ($store.currentTags.includes(tag)) {
      store.deselectTag(tag);
      return;
    }
    // otherwise, add it
    store.selectTag(tag);
  }
</script>

<ContextMenu
  menuItems={contextMenuItems}
  boundingElement={contextMenuBounds}
  class="inline"
>
  <TagChip {tag} clickAction={toggleTagSelection}></TagChip>
</ContextMenu>

<Modal bind:open={editModalOpen}>
  <h1>Edit tag</h1>
  <form on:submit|preventDefault={submitEdit}>
    <p>Id: {editedTag.id}</p>
    <label for="tag-name">Name:</label>
    <input type="text" id="tag-name" bind:value={editedTag.name} />

    <TagEditor
      tags={editedTag.parents}
      onAdd={addTagTag}
      onRemove={removeTagTag}
    ></TagEditor>

    <button on:click={cancelEdit}>Cancel</button>
    <button type="submit">Submit</button>
  </form>
</Modal>
