<script lang="ts">
  import TagEditor from "./lib/TagEditor.svelte";
  import store from "./lib/store";
  import type { tagger } from "./lib/wailsjs/go/models";

  export let file: tagger.File | undefined;

  // TODO: This might be leaving a trail of unused video players...
  let videoPlayer: HTMLMediaElement = document.createElement("video");

  async function addTag(tag: tagger.Tag) {
    if (!file) return;
    await store.tagFile(file, tag);
  }

  async function removeTag(tag: tagger.Tag) {
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
  $: if (file && file.filetype) {
    console.log(file.id);
    console.log(videoPlayer.canPlayType("video/" + file.filetype));
  }
</script>

{#if file === undefined}
  <p>No file selected</p>
{:else}
  {#key file.id}
    <!-- Preview file -->
    {#if imageFormats.includes(file.filetype)}
      <img src="/file/{file.id}" alt={file.hash} />
    {:else if videoFormats.includes(file.filetype)}
      <!-- svelte-ignore a11y-media-has-caption -->
      <video controls autoplay>
        <source src="/file/{file.id}" type="video/{file.filetype}" />
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
      <button on:click={removeFile}>Remove</button>
    </p>
  {/key}
{/if}
