/* eslint-disable @typescript-eslint/no-explicit-any */
import { Err, Error, ErrorCode, newError, Result } from "../util"

type JobCall<T> = () => Promise<Result<T, Error>>
type JobResolver<T> = (result: T) => void
type JobRejector = (error: Error) => void
type JobAfter<T> = (res: Result<T, Error>) => void

/**
 * A job that exposes a promise that can be resolved or rejected.
 */
export class Job<T> {
    public call: JobCall<T>
    public resolver?: JobResolver<T>
    public rejector?: JobRejector
    public after?: JobAfter<T>

    constructor(call: JobCall<T>, config: { resolver?: JobResolver<T>, rejector?: JobRejector, after?: JobAfter<T> }) {
        this.call = call
        this.resolver = config.resolver
        this.rejector = config.rejector
        this.after = config.after
    }
}

/**
 * A job that can be queueable with a attempts counter that rejects all run when reachs max attempts
 */
class QueableJob<T> extends Job<T> {

    private attempts: number
    private maxAttempts: number

    constructor(call: JobCall<T>, config: { resolver?: JobResolver<T>, rejector?: JobRejector, after?: JobAfter<T>, maxAttempts?: number }) {
        super(call, { resolver: config.resolver, rejector: config.rejector, after: config.after })
        this.attempts = 0
        this.maxAttempts = config.maxAttempts || 3
    }

    private incrementAttempts(): void {
        this.attempts++
    }

    public async run(): Promise<Result<any, Error>> {
        if (!this.stillValid()) {
            return Err(newError(ErrorCode.EMAXATTEMPTS, `Max attempts reached for job ${this.call.name}`))
        }
        this.incrementAttempts()
        return await this.call()
    }

    public stillValid(): bool {
        return this.attempts < this.maxAttempts
    }
}

/**
 * A queue manager that manages a queue of jobs.
 * It can be used to queue jobs and run them sequentially.
 * Expose a BehaviorMap to manage specific errors, like JWT expired.
 */
class QueueManager {

    private queue: Array<QueableJob<any>>
    private startedDispatching: bool
    private behaviorMap: Map<ErrorCode, Job<any>> = new Map()

    constructor() {
        this.queue = []
        this.startedDispatching = false
    }

    public add<T>(call: JobCall<T>, resolver?: JobResolver<T>, rejector?: JobRejector, after?: JobAfter<T>) {
        this.queue.push(new QueableJob(call, { resolver, rejector, after }))
        if (!this.startedDispatching) {
            this.dispatch()
        }
    }

    public flush() {
        this.queue = []
    }

    public registerBehavior<T>(errorCode: ErrorCode, call: Job<T>) {
        this.behaviorMap.set(errorCode, call)
    }

    private async dispatch() {

        if (this.queue.length === 0) {
            this.startedDispatching = false
            return
        }

        this.startedDispatching = true

        const job = this.queue.shift()

        if (job) {
            await this.dispatchSingleJob(job)
        }

        this.dispatch()
    }

    private async dispatchSingleJob<T>(queuedJob: QueableJob<T>) {

        const result = await queuedJob.run()

        if (result.isOk()) {

            if (queuedJob.resolver) {
                queuedJob.resolver(result.unwrap())
            }

        } else {

            const behavior = this.behaviorMap.get(result.unwrapErr().code)

            if (behavior) {

                const result = await behavior.call()
                if (result.isOk()) {
                    if (behavior.resolver) {
                        behavior.resolver(result.unwrap())
                    }
                    await this.dispatchSingleJob(queuedJob)
                } else {
                    if (behavior.rejector) {
                        behavior.rejector(result.unwrapErr())
                    }
                }

            } else if (queuedJob.rejector) {

                queuedJob.rejector(result.unwrapErr())
            }
        }

        if (queuedJob.after) {
            queuedJob.after(result)
        }
    }
}

const queueManager = new QueueManager()

export default queueManager