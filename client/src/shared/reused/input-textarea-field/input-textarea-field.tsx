import { useMemo } from 'react'
import { Control, FieldPath, FieldValues, RegisterOptions } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { FormField } from '@shared/reused/form-field/form-field.tsx'
import { InputTextarea } from '@shared/uikit/input-textarea'

interface InputTextareaFieldProps<
	TFieldValues extends FieldValues = FieldValues,
	TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
> {
	control: Control<TFieldValues>
	name: TName
	label?: string
	height?: number
	placeholder?: string
	required?: boolean
	rules?: RegisterOptions<TFieldValues, TName>
	onBlur?: (value: string) => void
}

export const InputTextareaField = <
	TFieldValues extends FieldValues = FieldValues,
	TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
>({
	control,
	name,
	label,
	height,
	placeholder,
	required = false,
	rules,
	onBlur: customOnBlur,
}: InputTextareaFieldProps<TFieldValues, TName>) => {
	const { t } = useTranslation()

	const defaultRules = useMemo(() => {
		return required ? { required: t('validation_error.required_field'), ...rules } : rules
	}, [required, rules, t])

	return (
		<FormField control={control} name={name} rules={defaultRules}>
			{({ value, onChange, onBlur, invalid, error }) => (
				<InputTextarea
					value={value || ''}
					label={label}
					height={height}
					placeholder={placeholder}
					required={required}
					invalid={invalid}
					error={error}
					onChange={onChange}
					onBlur={() => {
						const trimmedValue = value?.trim() || ''
						if (trimmedValue !== value) {
							onChange(trimmedValue)
						}
						if (customOnBlur) {
							customOnBlur(trimmedValue)
						}
						onBlur()
					}}
				/>
			)}
		</FormField>
	)
}
