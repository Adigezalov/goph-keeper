import { Password, PasswordProps } from 'primereact/password'
import { classNames } from 'primereact/utils'

import styles from './input-password.module.sass'

type Props = PasswordProps & {
	label?: string
	error?: string
	wrapperClassName?: string
	labelClassNamer?: string
	inputClassName?: string
}

export const InputPassword = ({
	label,
	error,
	wrapperClassName,
	labelClassNamer,
	inputClassName,
	...props
}: Props) => {
	return (
		<div className={classNames(styles.root, wrapperClassName)}>
			{label && (
				<span className={classNames(styles.label, labelClassNamer)}>
					{label}
					{props.required && <span className={styles.required}>*</span>}
				</span>
			)}
			<Password
				{...props}
				className={classNames(styles.input, inputClassName)}
				feedback={false}
			/>
			{error && <span className={styles.error}>{error}</span>}
		</div>
	)
}
