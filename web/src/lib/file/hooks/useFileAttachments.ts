import { computed, ref } from 'vue'
import { AttachmentManager } from '../attachments/attachment-manager'
import type { FileAttachment, FileValidationRule } from '../types'

export function useFileAttachments(initial: FileAttachment[] = []) {
  const manager = new AttachmentManager()
  manager.reset(initial)
  const revision = ref(0)

  const touch = () => {
    revision.value += 1
  }

  const files = computed(() => {
    revision.value
    return manager.list()
  })

  const addAttachment = (file: FileAttachment) => {
    manager.add(file)
    touch()
  }
  const addAttachments = (items: FileAttachment[]) => {
    manager.addMany(items)
    touch()
  }
  const removeAttachment = (index: number) => {
    manager.remove(index)
    touch()
  }
  const removeAttachmentById = (id: string) => {
    manager.removeById(id)
    touch()
  }
  const reorderAttachments = (from: number, to: number) => {
    manager.reorder(from, to)
    touch()
  }
  const resetAttachments = (items: FileAttachment[] = []) => {
    manager.reset(items)
    touch()
  }
  const validateAttachments = (rule: FileValidationRule = {}) => manager.validate(rule)

  return {
    files,
    addAttachment,
    addAttachments,
    removeAttachment,
    removeAttachmentById,
    reorderAttachments,
    resetAttachments,
    validateAttachments,
  }
}
