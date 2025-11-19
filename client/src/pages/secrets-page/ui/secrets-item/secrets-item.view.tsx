import { Control, FieldValues, Path } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { FileUploadField } from '@shared/reused/file-upload-field/file-upload-field.tsx'
import { InputPasswordField } from '@shared/reused/input-password-field'
import { InputTextField } from '@shared/reused/input-text-field'
import { Button } from '@shared/uikit/button'
import { Icon } from '@shared/uikit/icon'

import styles from './secrets-item.module.sass'

type TSecretFormFields = {
	login: string
	password: string
	metadata?: Record<string, string>
	binaryData?: File | Uint8Array
}

type Props<T extends FieldValues & TSecretFormFields> = {
	control: Control<T>
	onSave: () => void
	onDelete?: () => void
	onDownload?: () => void
	isEditMode?: boolean
	disabled?: boolean
}

export const SecretsItemView = <T extends FieldValues & TSecretFormFields>({
	control,
	onSave,
	onDelete,
	onDownload,
	isEditMode = false,
	disabled = false,
}: Props<T>) => {
	const { t } = useTranslation()

	const saveButtonIcon = isEditMode ? Icon.SAVE : Icon.PLUS

	return (
		<div className={styles.root}>
			<InputTextField<T>
				control={control}
				name={'metadata.app' as Path<T>}
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
			<FileUploadField
				control={control}
				name={'binaryData' as Path<T>}
				label={t('secrets.select_file')}
			/>
			<div className={styles.actions}>
				<Button
					onClick={onSave}
					icon={<i className={saveButtonIcon} />}
					disabled={disabled}
				/>
				<Button
					severity="info"
					icon={<i className="pi pi-download" />}
					onClick={onDownload}
					disabled={!onDownload}
				/>
				<Button
					severity="danger"
					icon={<i className={Icon.TRASH} />}
					onClick={onDelete}
					disabled={!onDelete}
				/>
			</div>
		</div>
	)
}
