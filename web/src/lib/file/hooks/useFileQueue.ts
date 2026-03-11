import { computed } from 'vue'
import { FileQueue, globalFileQueue } from '../queue/file-queue'
import type { FileUploadInput, QueueTask } from '../types'

const scopedQueues = new Map<string, FileQueue>()

function getQueue(scopeKey = 'global') {
  if (scopeKey === 'global') return globalFileQueue
  if (!scopedQueues.has(scopeKey)) {
    scopedQueues.set(scopeKey, new FileQueue())
  }
  return scopedQueues.get(scopeKey) as FileQueue
}

export function useFileQueue(scopeKey = 'global') {
  const queue = getQueue(scopeKey)

  const enqueueUpload = (input: FileUploadInput) => queue.enqueue(input)
  const enqueueUploads = (inputs: FileUploadInput[]) => queue.enqueueMany(inputs)
  const cancelUpload = (taskId: string) => queue.cancel(taskId)
  const clearFinishedUploads = () => queue.clearFinished()

  const tasks = computed<QueueTask[]>(() => queue.tasks.value)
  const doneItems = computed(() => queue.tasks.value.filter((task) => task.status === 'success'))
  const failedItems = computed(() => queue.tasks.value.filter((task) => task.status === 'failed'))
  const waitForTask = (taskId: string) =>
    new Promise<QueueTask>((resolve, reject) => {
      const timer = setInterval(() => {
        const task = queue.tasks.value.find((item) => item.id === taskId)
        if (!task) return
        if (task.status === 'success') {
          clearInterval(timer)
          resolve(task)
          return
        }
        if (['failed', 'cancelled'].includes(task.status)) {
          clearInterval(timer)
          reject(new Error(task.error || '上传任务失败'))
        }
      }, 60)
    })

  return {
    tasks,
    running: queue.running,
    doneItems,
    failedItems,
    enqueueUpload,
    enqueueUploads,
    waitForTask,
    cancelUpload,
    clearFinishedUploads,
  }
}
