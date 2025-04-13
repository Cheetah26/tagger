import { TaggerService } from "$bindings/index";
import { TagOrdering, type File, type Tag, type TagID, type TagMap } from "$bindings/pkg/tagger";

class AppState {
  allTags = $state({} as TagMap);
  selectedTags = $state([] as Tag[]);

  tagOrdering = $state(TagOrdering.FileCount);
  tagIdsOrdered = $state([] as TagID[]);

  currentFiles = $state([] as File[]);
  selectedFile = $state(undefined as File | undefined);

  async getFiles() {
    if (this.selectedTags.length == 0) {
      this.currentFiles = await TaggerService.GetAllFiles()
    } else {
      this.currentFiles = await TaggerService.GetFilesByTag(this.selectedTags);
    }

    this.selectedFile = this.currentFiles.find(f => f.id == this.selectedFile?.id);
  }
}

export let appState = new AppState();