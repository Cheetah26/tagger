<script lang="ts">
  import { TaggerService } from "$bindings/index";
  import type { File, Tag } from "$bindings/pkg/tagger";
  import { appState } from "$lib/state.svelte";
  import TagEditor from "$lib/TagEditor.svelte";

  let { file }: { file: File } = $props();

  let newDescription = $state(file.description);

  async function open() {
    TaggerService.OpenFile(file);
  }

  async function reveal() {
    TaggerService.Reveal(file);
  }

  async function remove() {
    if (confirm("Are you sure?")) {
      await TaggerService.RemoveFile(file);
      appState.getFiles();
    }
  }

  async function editDescription() {
    await TaggerService.UpdateFileDescription({
      ...file,
      description: newDescription,
    });
    appState.getFiles();
  }

  async function addTag(tag: Tag) {
    await TaggerService.TagFile(file, tag);
    appState.getFiles();
  }

  async function removeTag(tag: Tag) {
    await TaggerService.UntagFile(file, tag);
    appState.getFiles();
  }
</script>

<div class="h-full overflow-y-auto p-1">
  <div class="flex flex-row mb-1">
    {#snippet button(text: string, action: () => void)}
      <button class="border mx-1 p-1" onclick={action}>{text}</button>
    {/snippet}

    {@render button("Open", open)}
    {@render button("Reveal", reveal)}
    <span class="grow"></span>
    {@render button("Remove", remove)}
  </div>

  <div class="relative">
    <p class="flex flex-row justify-between">
      <span>Description:</span><span
        >Id: {file.id} | Filetype: {file.filetype}</span
      >
    </p>
    <textarea
      class="w-full h-full overflow-y-scroll"
      bind:value={newDescription}
      contenteditable="plaintext-only"
    ></textarea>
    {#if newDescription != file.description}
      <button
        onclick={editDescription}
        class="absolute bottom-3 right-2 p-1 border">Save</button
      >
    {/if}
  </div>

  <TagEditor tags={file.tags} onAdd={addTag} onRemove={removeTag}></TagEditor>
</div>
