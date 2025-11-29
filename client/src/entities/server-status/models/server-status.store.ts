import { makeAutoObservable, runInAction } from 'mobx'

import { TStoreLogic } from '@shared/store'

import { serverStatusApi } from '../api'

const CHECK_INTERVAL_MS = 30000

export class ServerStatusStore {
	status = false
	private intervalId: ReturnType<typeof setInterval> | null = null

	rootStore: TStoreLogic

	constructor(rootStore: TStoreLogic) {
		this.rootStore = rootStore
		makeAutoObservable(this, {}, { autoBind: true })
	}

	checkStatus = async () => {
		try {
			const response = await serverStatusApi()

			if (response.status === 200) {
				runInAction(() => {
					this.status = true
				})
			}
		} catch (_) {
			runInAction(() => {
				this.status = false
			})
		}
	}

	startStatusCheck = () => {
		this.stopStatusCheck()

		void this.checkStatus()

		this.intervalId = setInterval(() => {
			void this.checkStatus()
		}, CHECK_INTERVAL_MS)
	}

	stopStatusCheck = () => {
		if (this.intervalId) {
			clearInterval(this.intervalId)
			this.intervalId = null
		}
	}

	clearServerStatusStore = () => {
		this.stopStatusCheck()
		this.status = false
	}
}
