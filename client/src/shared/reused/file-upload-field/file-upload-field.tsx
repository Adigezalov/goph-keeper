import { FileUpload as PrimeFileUpload } from 'primereact/fileupload'
import { useEffect, useMemo, useRef } from 'react'
import { Control, FieldPath, FieldValues, RegisterOptions } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { FileUpload } from '@shared/uikit/file-upload'

import { FormField } from '../form-field'

interface InputTextFieldProps<
	TFieldValues extends FieldValues = FieldValues,
	TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
> {
	control: Control<TFieldValues>
	name: TName
	label?: string
	placeholder?: string
	required?: boolean
	disabled?: boolean
	rules?: RegisterOptions<TFieldValues, TName>
}

type FileUploadComponentProps = {
	value: any
	onChange: (file: File) => void
	label?: string
	disabled?: boolean
	fileUploadRef: React.RefObject<PrimeFileUpload | null>
}

const FileUploadComponent = ({
	value,
	onChange,
	label,
	disabled,
	fileUploadRef,
}: FileUploadComponentProps) => {
	useEffect(() => {
		if (!(value instanceof File) && fileUploadRef.current) {
			fileUploadRef.current.clear()
		}
	}, [value, fileUploadRef])

	return (
		<FileUpload
			ref={fileUploadRef}
			chooseLabel={label}
			disabled={disabled}
			onSelect={(event) => {
				const file = event.files[0]
				if (file) {
					onChange(file)
				}
			}}
		/>
	)
}

export const FileUploadField = <
	TFieldValues extends FieldValues = FieldValues,
	TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
>({
	control,
	name,
	label,
	required = false,
	disabled = false,
	rules,
}: InputTextFieldProps<TFieldValues, TName>) => {
	const { t } = useTranslation()
	const fileUploadRef = useRef<PrimeFileUpload>(null)

	const defaultRules = useMemo(() => {
		return required ? { required: t('validation_error.required_field'), ...rules } : rules
	}, [required, rules, t])

	return (
		<FormField control={control} name={name} rules={defaultRules}>
			{({ value, onChange }) => (
				<FileUploadComponent
					value={value}
					onChange={onChange}
					label={label}
					disabled={disabled}
					fileUploadRef={fileUploadRef}
				/>
			)}
		</FormField>
	)
}
