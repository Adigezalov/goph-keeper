import { AuthStore } from '@entities/auth/models'
import { CryptoStore } from '@entities/crypto/models'
import { NetworkStatusStore } from '@entities/network-status/models'
import { ServerStatusStore } from '@entities/server-status/models'

import { SecretsPageStore } from '@pages/secrets-page/models'

export type TStoreLogic = {
	auth: AuthStore
	cryptoKey: CryptoStore
	networkStatus: NetworkStatusStore
	secretsPage: SecretsPageStore
	serverStatus: ServerStatusStore
}
