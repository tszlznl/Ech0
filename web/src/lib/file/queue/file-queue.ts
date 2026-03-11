import { ref } from 'vue'
import { fileApiAdapter } from '../api/adapter'
import type { FileApiAdapter } from '../api/adapter'
import type { QueueOptions, QueueTask, QueueTaskStatus } from '../types'

type TaskInternal = QueueTask & { cancelled?: boolean }

function nextStatus(task: TaskInternal, status: QueueTaskStatus, error?: string) {
  task.status = status
  task.error = error
}

export class FileQueue {
  readonly tasks = ref<TaskInternal[]>([])
  readonly running = ref(false)

  private options: Required<QueueOptions>
  private adapter: FileApiAdapter

  constructor(options: QueueOptions = {}, adapter = fileApiAdapter) {
    this.options = {
      concurrency: options.concurrency ?? 2,
      maxRetry: options.maxRetry ?? 2,
    }
    this.adapter = adapter
  }

  enqueue(input: QueueTask['input']) {
    const task: TaskInternal = {
      id: `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`,
      name: input.file.name,
      status: 'queued',
      attempt: 0,
      progress: 0,
      input,
    }
    this.tasks.value.push(task)
    void this.run()
    return task.id
  }

  enqueueMany(inputs: QueueTask['input'][]) {
    return inputs.map((input) => this.enqueue(input))
  }

  cancel(taskId: string) {
    const task = this.tasks.value.find((item) => item.id === taskId)
    if (!task) return
    task.cancelled = true
    if (task.status === 'queued' || task.status === 'retrying') {
      nextStatus(task, 'cancelled')
    }
  }

  clearFinished() {
    this.tasks.value = this.tasks.value.filter((task) =>
      ['queued', 'running', 'retrying'].includes(task.status),
    )
  }

  private async run() {
    if (this.running.value) return
    this.running.value = true
    try {
      while (true) {
        const pending = this.tasks.value.filter((task) => task.status === 'queued')
        const active = this.tasks.value.filter((task) => task.status === 'running').length
        if (!pending.length && active === 0) break

        const available = Math.max(this.options.concurrency - active, 0)
        if (available > 0) {
          const batch = pending.slice(0, available)
          await Promise.all(batch.map((task) => this.execute(task)))
        } else {
          await new Promise((resolve) => setTimeout(resolve, 40))
        }
      }
    } finally {
      this.running.value = false
    }
  }

  private async execute(task: TaskInternal) {
    if (task.cancelled) {
      nextStatus(task, 'cancelled')
      return
    }

    task.attempt += 1
    task.progress = 20
    nextStatus(task, 'running')
    try {
      const result = await this.adapter.uploadFile(task.input)
      if (task.cancelled) {
        nextStatus(task, 'cancelled')
        return
      }
      task.progress = 100
      task.result = result
      nextStatus(task, 'success')
    } catch (err) {
      const msg = err instanceof Error ? err.message : '上传失败'
      if (task.attempt <= this.options.maxRetry && !task.cancelled) {
        nextStatus(task, 'retrying', msg)
        await new Promise((resolve) => setTimeout(resolve, 250))
        nextStatus(task, 'queued')
        return
      }
      nextStatus(task, task.cancelled ? 'cancelled' : 'failed', msg)
    }
  }
}

export const globalFileQueue = new FileQueue()
