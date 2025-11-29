import { InputTextProps, InputText as PrimeInputText } from 'primereact/inputtext'
import { classNames } from 'primereact/utils'

import styles from './input-text.module.sass'

type Props = InputTextProps & {
	label?: string
	error?: string
	wrapperClassName?: string
	labelClassNamer?: string
	inputClassName?: string
}

export const InputText = ({
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
			<PrimeInputText {...props} className={classNames(styles.input, inputClassName)} />
			{error && <span className={styles.error}>{error}</span>}
		</div>
	)
}
