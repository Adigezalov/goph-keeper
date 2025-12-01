import { makeAutoObservable, runInAction } from 'mobx'

import { TStoreLogic } from '@shared/store'

export class NetworkStatusStore {
	isOnline = navigator.onLine

	rootStore: TStoreLogic

	constructor(rootStore: TStoreLogic) {
		this.rootStore = rootStore
		makeAutoObservable(this, {}, { autoBind: true })
	}

	private handleOnline = () => {
		runInAction(() => {
			this.isOnline = true
		})
	}

	private handleOffline = () => {
		runInAction(() => {
			this.isOnline = false
		})
	}

	initNetworkStatusStore = () => {
		runInAction(() => {
			this.isOnline = navigator.onLine
		})

		window.addEventListener('online', this.handleOnline)
		window.addEventListener('offline', this.handleOffline)
	}

	clearNetworkStatusStore = () => {
		window.removeEventListener('online', this.handleOnline)
		window.removeEventListener('offline', this.handleOffline)

		this.isOnline = true
	}
}
