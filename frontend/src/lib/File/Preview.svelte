<script lang="ts">
  import type { File } from "$bindings/pkg/tagger";

  let { file }: { file: File } = $props();
  let filePath = $derived(`/file/${file.id}?hash=${file.hash}`);

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

{#if imageFormats.includes(file.filetype)}
  <!-- svelte-ignore a11y_missing_attribute -->
  <img src={filePath} class="w-full h-full object-contain" />
{:else if videoFormats.includes(file.filetype)}
  <!-- svelte-ignore a11y_media_has_caption -->
  <video controls autoplay class="w-full h-full object-contain">
    <source src={filePath} type="video/{file.filetype}" />
  </video>
{:else}
  <div class="flex w-full h-full justify-center items-center">
    <p class="bg-orange-300 p-3">
      unable to preview filetype: {file.filetype}
    </p>
  </div>
{/if}
