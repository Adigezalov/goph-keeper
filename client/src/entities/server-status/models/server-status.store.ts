import { makeAutoObservable, runInAction } from 'mobx'

import { TStoreLogic } from '@shared/store'

import { serverStatusApi } from '../api'

const CHECK_INTERVAL_MS = 30000 // 30 секунд

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

	// Запускает периодическую проверку статуса сервера каждые 30 секунд
	startStatusCheck = () => {
		// Останавливаем предыдущий интервал, если он есть
		this.stopStatusCheck()

		// Первая проверка сразу
		void this.checkStatus()

		// Запускаем интервал
		this.intervalId = setInterval(() => {
			void this.checkStatus()
		}, CHECK_INTERVAL_MS)
	}

	// Останавливает периодическую проверку статуса сервера
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
