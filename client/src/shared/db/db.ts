import Dexie, { Table } from 'dexie'

import { TSecret, TSyncQueue } from '@pages/secrets-page/types'

export class KeeperDB extends Dexie {
	secrets!: Table<TSecret, string>
	syncQueue!: Table<TSyncQueue, string>

	constructor() {
		super('KeeperDB')

		this.version(2).stores({
			secrets: 'localId, id, serverId, syncStatus, updatedAt, deletedAt',
			syncQueue: 'id, timestamp, action, secretId',
		})
	}

	// Метод для полной очистки базы данных
	async clearAllData() {
		await this.secrets.clear()
		await this.syncQueue.clear()
	}

	// Метод для полного удаления и пересоздания базы данных
	async resetDatabase() {
		await this.delete()
		await this.open()
	}
}

export const db = new KeeperDB()

// Глобальная функция для очистки (доступна в консоли браузера)
if (typeof window !== 'undefined') {
	;(window as any).clearKeeperDB = async () => {
		await db.clearAllData()
		console.log('✅ База данных очищена')
		window.location.reload()
	}
	;(window as any).resetKeeperDB = async () => {
		await db.resetDatabase()
		console.log('✅ База данных удалена и пересоздана')
		window.location.reload()
	}
}
