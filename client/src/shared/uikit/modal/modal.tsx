import { Dialog, DialogProps } from 'primereact/dialog'
import { classNames } from 'primereact/utils'

import styles from './modal.module.sass'

export const Modal = (props: DialogProps) => {
	return (
		<Dialog
			{...props}
			className={classNames(styles.root, props.className)}
			headerClassName={classNames(styles.header, props.headerClassName)}
			contentClassName={classNames(styles.content, props.contentClassName)}
		>
			{props.children}
		</Dialog>
	)
}
