<script lang="ts">
  import TagEditor from "./TagEditor.svelte";
  import store from "./store";
  import type { Tag } from "../../bindings/github.com/cheetah26/tagger/pkg/tagger";

  $: file = $store.currentFile;
  $: filePath = file && `/file/${file.id}`

  async function addTag(tag: Tag) {
    if (!file) return;
    await store.tagFile(file, tag);
  }

  async function removeTag(tag: Tag) {
    if (!file) return;
    await store.untagFile(file, tag);
  }

  async function removeFile() {
    if (file && confirm("Are you sure?")) {
      await store.removeFile(file);
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
    <!-- Preview file -->
    {#if imageFormats.includes(file.filetype)}
      <img src={filePath} alt={file.hash} />
    {:else if videoFormats.includes(file.filetype)}
      <!-- svelte-ignore a11y-media-has-caption -->
      <video controls autoplay>
        <source src={filePath} type="video/{file.filetype}" />
      </video>
    {:else}
      <p class="bg-orange-500">Format {file.filetype} not supported</p>
    {/if}

    <p class="break-all">{file.hash.slice(0, 8)}</p>

    <!-- Tags -->
    <TagEditor tags={file.tags} onAdd={addTag} onRemove={removeTag}></TagEditor>

    <p>Description:</p>
    {#if file.description}
      <p class="break-all">{file.description}</p>
      <p>HI</p>
    {/if}

    <p class="mt-4">
      <button on:click={store.openCurrentFile}>Open</button>
      <button on:click={store.revealCurrentFile}>Reveal</button>
      <button on:click={removeFile}>Remove</button>
    </p>
  {/key}
{/if}
