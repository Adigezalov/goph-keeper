import { Divider } from 'primereact/divider'

import { TSecret } from '../../types'
import { SecretsItem, SecretsItemNew } from '../secrets-item'

import styles from './secrets-page.module.sass'

type Props = {
	secrets: TSecret[]
}

export const SecretsPageView = ({ secrets }: Props) => {
	return (
		<div className={styles.root}>
			<SecretsItemNew />
			<Divider />
			{secrets.map((secret) => (
				<SecretsItem secret={secret} key={secret.localId} />
			))}
		</div>
	)
}
