import { Control, FieldValues, Path } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { InputPasswordField } from '@shared/reused/input-password-field'
import { InputTextField } from '@shared/reused/input-text-field'
import { Button } from '@shared/uikit/button'
import { Icon } from '@shared/uikit/icon'

import styles from './secrets-item.module.sass'

type TSecretFormFields = {
	login: string
	password: string
	metadata?: string
}

type Props<T extends FieldValues & TSecretFormFields> = {
	control: Control<T>
	onSave: () => void
	onDelete?: () => void
	isEditMode?: boolean
	disabled?: boolean
}

export const SecretsItemView = <T extends FieldValues & TSecretFormFields>({
	control,
	onSave,
	onDelete,
	isEditMode = false,
	disabled = false,
}: Props<T>) => {
	const { t } = useTranslation()

	const saveButtonIcon = isEditMode ? Icon.SAVE : Icon.PLUS

	return (
		<div className={styles.root}>
			<InputTextField<T>
				control={control}
				name={'metadata' as Path<T>}
				label={t('app')}
				required={!isEditMode || !disabled}
			/>
			<InputTextField<T>
				control={control}
				name={'login' as Path<T>}
				label={t('username')}
				required={!isEditMode || !disabled}
			/>
			<InputPasswordField<T>
				control={control}
				name={'password' as Path<T>}
				label={t('password')}
				required={!isEditMode || !disabled}
			/>
			<div className={styles.actions}>
				<Button
					onClick={onSave}
					icon={<i className={saveButtonIcon} />}
					disabled={disabled}
				/>
				{onDelete && (
					<Button
						severity="danger"
						icon={<i className={Icon.TRASH} />}
						onClick={onDelete}
					/>
				)}
			</div>
		</div>
	)
}
