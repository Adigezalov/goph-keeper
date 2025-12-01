import { classNames } from 'primereact/utils'

import { Icon } from '@shared/uikit/icon'

import styles from './server-status.module.sass'

type Props = {
	status: boolean
}

export const ServerStatusView = ({ status }: Props) => {
	return (
		<span className={classNames(Icon.SERVER, styles.icon, { [styles.error]: !status })} />
	)
}
