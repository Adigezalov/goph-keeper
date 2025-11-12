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
}

export const db = new KeeperDB()
