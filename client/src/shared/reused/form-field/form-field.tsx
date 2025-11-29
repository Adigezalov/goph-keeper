import { ReactElement } from 'react'
import { Control, Controller, FieldPath, FieldValues, RegisterOptions } from 'react-hook-form'

interface FormFieldProps<
	TFieldValues extends FieldValues = FieldValues,
	TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>
> {
	control: Control<TFieldValues>
	name: TName
	rules?: RegisterOptions<TFieldValues, TName>
	children: (props: {
		value: any
		onChange: (value: any) => void
		onBlur: () => void
		invalid: boolean
		error?: string
	}) => ReactElement
}

export const FormField = <
	TFieldValues extends FieldValues = FieldValues,
	TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>
>({
	control,
	name,
	rules,
	children,
}: FormFieldProps<TFieldValues, TName>) => {
	return (
		<Controller
			control={control}
			name={name}
			rules={rules}
			render={({ field: { value, onChange, onBlur }, fieldState: { invalid, error } }) =>
				children({
					value,
					onChange,
					onBlur,
					invalid,
					error: error?.message,
				})
			}
		/>
	)
}