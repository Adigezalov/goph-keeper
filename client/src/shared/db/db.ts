import i18next from 'i18next'
import Dexie, { Table } from 'dexie'

import { TSecret, TSyncQueue } from '@pages/secrets-page/types'
import { TSyncMeta } from './db.types'

export class KeeperDB extends Dexie {
	secrets!: Table<TSecret, string>
	syncQueue!: Table<TSyncQueue, string>
	syncMeta!: Table<TSyncMeta, string>

	constructor() {
		super('KeeperDB')

		this.version(3).stores({
			secrets: 'localId, id, serverId, syncStatus, updatedAt, deletedAt',
			syncQueue: 'id, timestamp, action, secretId',
			syncMeta: 'key',
		})
	}

	async clearAllData() {
		await this.secrets.clear()
		await this.syncQueue.clear()
		await this.syncMeta.clear()
	}

	async resetDatabase() {
		await this.delete()
		await this.open()
	}
}

export const db = new KeeperDB()

if (typeof window !== 'undefined') {
	;(window as any).clearKeeperDB = async () => {
		await db.clearAllData()
		console.log(i18next.t('db.cleared'))
		window.location.reload()
	}
	;(window as any).resetKeeperDB = async () => {
		await db.resetDatabase()
		console.log(i18next.t('db.reset'))
		window.location.reload()
	}
}
