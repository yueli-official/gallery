let draftSequence = 0;

// Upload queue keys exist only in one browser draft and never cross the API.
// Persisted Gallery entities receive their IDs from the Go Identifier module.
export function galleryDraftKey(): string {
  draftSequence += 1;
  return `draft-gallery-upload-${draftSequence}`;
}
