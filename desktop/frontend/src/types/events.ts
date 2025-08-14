// Event type definitions for React components

import { ChangeEvent, FormEvent, MouseEvent, KeyboardEvent, FocusEvent } from 'react';

// Input events
export type InputChangeEvent = ChangeEvent<HTMLInputElement>;
export type TextAreaChangeEvent = ChangeEvent<HTMLTextAreaElement>;
export type SelectChangeEvent = ChangeEvent<HTMLSelectElement>;

// Form events
export type FormSubmitEvent = FormEvent<HTMLFormElement>;

// Mouse events
export type ButtonClickEvent = MouseEvent<HTMLButtonElement>;
export type DivClickEvent = MouseEvent<HTMLDivElement>;
export type AnchorClickEvent = MouseEvent<HTMLAnchorElement>;

// Keyboard events
export type InputKeyboardEvent = KeyboardEvent<HTMLInputElement>;

// Focus events
export type InputFocusEvent = FocusEvent<HTMLInputElement>;

// Generic handler types
export type ChangeHandler<T = HTMLInputElement> = (event: ChangeEvent<T>) => void;
export type SubmitHandler = (event: FormSubmitEvent) => void;
export type ClickHandler<T = HTMLButtonElement> = (event: MouseEvent<T>) => void;