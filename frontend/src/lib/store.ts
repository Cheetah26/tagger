import { get, writable } from "svelte/store";
import { TaggerService } from "../../bindings/github.com/cheetah26/tagger";
import { File, Tag, type TagMap } from "../../bindings/github.com/cheetah26/tagger/pkg/tagger";
import { GetTagByName } from "../../bindings/github.com/cheetah26/tagger/taggerservice";

type StoreContents = {
  files: File[];
  tags: TagMap;

  currentFile?: File;
  currentTags: Tag[];
}

function CreateStore() {
  const emptyStore = {
    files: [],
    tags: [],

    currentFile: undefined,
    currentTags: [],
  }
  const store = writable<StoreContents>(emptyStore)
  const { subscribe, set, update } = store

  async function open() {
    await TaggerService.Open()

    set(emptyStore)
    await getAllTags()
    await getFiles()
  }

  async function getAllTags() {
    const newTags = await TaggerService.GetAllTags()

    update(s => {
      // keep the user's tag selection
      let newCurrentTags: Tag[] = [];
      for (let ct of s.currentTags) {
        let nct = Object.keys(newTags).find(id => Number(id) == ct.id)
        if (nct) {
          newCurrentTags.push(newTags[nct as `${number}`]);
        }
      }

      return {
        ...s,
        tags: newTags,
        currentTags: newCurrentTags
      } as StoreContents
    })
  }

  async function getFiles() {
    const state = get(store)

    if (state.currentTags.length == 0) {
      state.files = await TaggerService.GetAllFiles()
    } else {
      state.files = await TaggerService.GetFilesByTag(state.currentTags)
    }

    // Deselect the current file if it no longer meets the filter
    if (state.currentFile && state.files && !state.files.some(f => f.hash == state.currentFile?.hash)) {
      state.currentFile = undefined
    }

    set(state)
  }

  async function selectFile(file: File) {
    // const fullFile = await TaggerService.GetFile(file.id)
    update(s => ({
      ...s,
      currentFile: file
    }))
  }

  async function importFiles() {
    await TaggerService.ImportFilesDialog()
    await getFiles()
  }

  async function selectTag(tag: Tag) {
    update(s => {
      s.currentTags.push(tag)
      return s
    })

    await getFiles()
  }

  async function deselectTag(tag: Tag) {
    update(s => {
      s.currentTags = s.currentTags.filter(t => t.id != tag.id)
      return s
    })
    await getFiles()
  }

  async function removeFile(file: File) {
    await TaggerService.RemoveFile(file)
    update(s => {
      s.currentFile = undefined
      return s
    })
    await getFiles()
  }

  async function addTag(name: string): Promise<Tag | null> {
    const newTag = await TaggerService.AddTag(name).catch(e => console.error(e))
    await getAllTags()

    return newTag || null
  }

  async function tagFile(file: File, tag: Tag) {
    await TaggerService.TagFile(file, tag)
    await selectFile(file)
    await getAllTags()
  }

  async function untagFile(file: File, tag: Tag) {
    await TaggerService.UntagFile(file, tag)
    await selectFile(file)
  }

  async function openCurrentFile() {
    const state = get(store)
    if (state.currentFile) {
      await TaggerService.OpenFile(state.currentFile)
    }
  }

  async function revealCurrentFile() {
    const state = get(store)
    if (state.currentFile) {
      await TaggerService.Reveal(state.currentFile)
    }
  }

  async function getUntaggedFiles() {
    const files = await TaggerService.GetUntaggedFiles()
    update(s => {
      s.files = files
      console.log(files[0])
      console.log(s.files[0])
      return s
    })
  }

  async function removeTag(tag: Tag) {
    await TaggerService.RemoveTag(tag)
    await getAllTags()

    // remove tag from currentTags selection
    update(s => {
      s.currentTags = s.currentTags.filter(t => t.id != tag.id);
      return s
    })

    // re-fetch files with the new current tags
    await getFiles()

    // update selected file
    const state = get(store)
    if (state.currentFile) {
      await selectFile(state.currentFile)
    }
  }


  async function updateTag(tag: Tag) {
    await TaggerService.UpdateTag(tag)
    await getAllTags()

    // update selected file
    const state = get(store)
    if (state.currentFile) {
      await selectFile(state.currentFile)
    }
  }

  return {
    subscribe,
    open,
    getFiles,
    getAllTags,
    selectFile,
    selectTag,
    deselectTag,
    removeFile,
    tagFile,
    untagFile,
    openCurrentFile,
    revealCurrentFile,
    getUntaggedFiles,
    removeTag,
    importFiles,
    updateTag,
    addTag
  }
}

const store = CreateStore()
export default store