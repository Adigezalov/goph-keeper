import {
	InputTextareaProps,
	InputTextarea as PrimeInputTextarea,
} from 'primereact/inputtextarea'
import { classNames } from 'primereact/utils'

import styles from './input-textarea.module.sass'

type Props = InputTextareaProps & {
	label?: string
	error?: string
	height?: number
	wrapperClassName?: string
	labelClassName?: string
	inputClassName?: string
}

export const InputTextarea = ({
	label,
	error,
	height = 200,
	wrapperClassName,
	labelClassName,
	inputClassName,
	...props
}: Props) => {
	return (
		<div className={classNames(styles.root, wrapperClassName)}>
			{label && (
				<span className={classNames(styles.label, labelClassName)}>
					{label}
					{props.required && <span className={styles.required}>*</span>}
				</span>
			)}
			<PrimeInputTextarea
				{...props}
				className={classNames(styles.input, inputClassName)}
				pt={{
					root: {
						style: {
							height: `${height}px`,
							minHeight: `${height}px`,
							resize: 'vertical',
						},
					},
				}}
			/>
			{error && <span className={styles.error}>{error}</span>}
		</div>
	)
}
