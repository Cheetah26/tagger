import { get, writable } from "svelte/store";
import { GetAllFiles, GetAllTags, GetFiles, Open, OpenDBDialog, ImportFile, GetFile, RemoveFile, UntagFile, TagFile, GetTag, AddTag, OpenFile, GetUntaggedFiles, RemoveTag, ImportFilesDialog, UpdateTag } from "./wailsjs/go/app/TaggerApp";
import { tagger } from "./wailsjs/go/models";
import { OnFileDrop } from "./wailsjs/runtime/runtime";

type StoreContents = {
  files: tagger.File[];
  tags: tagger.Tag[];

  currentFile?: tagger.File;
  currentTags: tagger.Tag[];
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
    const path = await OpenDBDialog()
    await Open(path)

    set(emptyStore)
    await getAllTags()
    await getFiles()
  }

  async function getAllTags() {
    const newTags = await GetAllTags()
    update(s => {
      let newCurrentTags = [];
      for (let ct of s.currentTags) {
        let nct = newTags.find(t => t.id == ct.id)
        if (nct) {
          newCurrentTags.push(nct)
        }
      }

      return {
        ...s,
        tags: newTags,
        currentTags: newCurrentTags
      }
    })
  }

  async function getFiles() {
    const state = get(store)

    if (state.currentTags.length == 0) {
      state.files = await GetAllFiles()
    } else {
      state.files = await GetFiles(state.currentTags)
    }

    // Deselect the current file if it no longer meets the filter
    if (state.currentFile && state.files && !state.files.some(f => f.hash == state.currentFile?.hash)) {
      state.currentFile = undefined
    }

    set(state)
  }

  async function selectFile(file: tagger.File) {
    const fullFile = await GetFile(file.id)
    update(s => ({
      ...s,
      currentFile: fullFile
    }))
  }

  OnFileDrop((_x, _y, paths) => {
    for (const path of paths) {
      ImportFile(path)
    }
    getFiles()
  }, false);

  async function importFiles() {
    await ImportFilesDialog()
    await getFiles()
  }

  async function selectTag(tag: tagger.Tag) {
    update(s => {
      s.currentTags.push(tag)
      return s
    })

    await getFiles()
  }

  async function deselectTag(tag: tagger.Tag) {
    update(s => {
      s.currentTags = s.currentTags.filter(t => t.id != tag.id)
      return s
    })
    await getFiles()
  }

  async function removeFile(file: tagger.File) {
    await RemoveFile(file)
    update(s => {
      s.currentFile = undefined
      return s
    })
    await getFiles()
  }

  async function addTag(name: string): Promise<tagger.Tag> {
    const newTag = await AddTag(name)
    await getAllTags()

    return newTag
  }

  async function tagFile(file: tagger.File, tag: tagger.Tag) {
    await TagFile(file, tag)
    await selectFile(file)
    await getAllTags()
  }

  async function untagFile(file: tagger.File, tag: tagger.Tag) {
    await UntagFile(file, tag)
    await selectFile(file)
  }

  async function openCurrentFile() {
    const state = get(store)
    if (state.currentFile) {
      await OpenFile(state.currentFile)
    }
  }

  async function getUntaggedFiles() {
    const files = await GetUntaggedFiles()
    update(s => {
      s.files = files
      console.log(files[0])
      console.log(s.files[0])
      return s
    })
  }

  async function removeTag(tag: tagger.Tag) {
    await RemoveTag(tag)
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


  async function updateTag(tag: tagger.Tag) {
    await UpdateTag(tag)
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
    getUntaggedFiles,
    removeTag,
    importFiles,
    updateTag,
    addTag
  }
}

const store = CreateStore()
export default store