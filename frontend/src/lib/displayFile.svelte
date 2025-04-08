<script lang="ts">
  import { TaggerService } from "$bindings/index";
  import type { Tag } from "$bindings/pkg/tagger";
  import TagEditor from "$lib/TagEditor.svelte";
  import { Pane, PaneGroup, PaneResizer } from "paneforge";
  import { appState } from "./state.svelte";

  let file = $derived(appState.selectedFile);
  let filePath = $derived(file && `/file/${file.id}`);

  async function addTag(tag: Tag) {
    if (!file) return;
    await TaggerService.TagFile(file, tag);
    appState.getFiles();
  }

  async function removeTag(tag: Tag) {
    if (!file) return;
    await TaggerService.UntagFile(file, tag);
    appState.getFiles();
  }

  async function removeFile() {
    if (file && confirm("Are you sure?")) {
      await TaggerService.RemoveFile(file);
      appState.getFiles();
    }
  }

  const imageFormats = [
    "apng",
    "avif",
    "gif",
    "jpg",
    "jpeg",
    "jfif",
    "pjpeg",
    "pjp",
    "png",
    "svg",
    "webp",
  ];
  const videoFormats = [
    "webm",
    "mkv",
    "flv",
    "ogg",
    "gifv",
    "avi",
    "mov",
    "mp4",
    "m4p",
    "flv",
  ];
</script>

{#if file === undefined}
  <p>No file selected</p>
{:else}
  {#key file.id}
    <PaneGroup direction="vertical">
      <Pane defaultSize={1}>
        <!-- Preview file -->
        {#if imageFormats.includes(file.filetype)}
          <!-- svelte-ignore a11y_missing_attribute -->
          <img src={filePath} class="w-full h-full object-contain" />
        {:else if videoFormats.includes(file.filetype)}
          <!-- svelte-ignore a11y_media_has_caption -->
          <video controls autoplay class="w-full h-full object-contain">
            <source src={filePath} type="video/{file.filetype}" />
          </video>
        {:else}
          <p class="bg-orange-500">Format {file.filetype} not supported</p>
        {/if}
      </Pane>
      <PaneResizer class="h-1 my-1 border"></PaneResizer>
      <Pane defaultSize={2}>
        <div class="h-full overflow-y-auto">
          <p class="break-all">{file.hash.slice(0, 8)}</p>

          <!-- Tags -->
          <TagEditor tags={file.tags} onAdd={addTag} onRemove={removeTag}
          ></TagEditor>

          <p>Description:</p>
          {#if file.description}
            <p class="break-all">{file.description}</p>
          {/if}

          <p class="mt-4">
            <button onclick={() => TaggerService.OpenFile(file)}>Open</button>
            <button onclick={() => TaggerService.Reveal(file)}>Reveal</button>
            <button onclick={removeFile}>Remove</button>
          </p>
        </div>
      </Pane>
    </PaneGroup>
  {/key}
{/if}
