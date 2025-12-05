// Type definitions for Airtable internal scripting,
// such as automations and extensions.

export interface Input {
    config(): { [key: string]: any },
}
// noinspection JSUnusedGlobalSymbols
export const input: Input

export interface Output {
    text(text: string): void,
}
// noinspection JSUnusedGlobalSymbols
export const output: Output

export interface Base {
    getTable(idOrName: string): Table,
}
// noinspection JSUnusedGlobalSymbols
export const base: Base

export function remoteFetchAsync(url: string, options: RequestInit): Promise<Response>

export interface Table {
    id: string,
    name: string,
    description?: string,
    url: string,
    fields: Field[],
    selectRecordAsync(
        recordId: string,
        options?: {
            fields?: Array<Field | string>,
        },
    ): Promise<Record | null>
    selectRecordsAsync(options?: {
        sorts?: Array<{
            field: Field | string,
            direction?: 'asc' | 'desc',
        }>,
        fields?: Array<Field | string>,
        recordIds?: Array<string>,
    }): Promise<RecordQueryResult>
    createRecordAsync(
        fields: {[fieldNameOrId: string]: unknown}
    ): Promise<string>
    createRecordsAsync(
        records: Array<{fields: {[fieldNameOrId: string]: unknown}}>
    ): Promise<Array<string>>
    updateRecordAsync(
        recordOrRecordId: Record | string,
        fields: {[fieldNameOrId: string]: unknown}
    ): Promise<void>
    updateRecordsAsync(records: Array<{
        id: string,
        fields: {[fieldNameOrId: string]: unknown},
    }>): Promise<void>
}

export interface Field {
    id: string,
    name: string,
    description?: string,
    type: string,
}

export interface Record {
    id: string,
    name: string,
    getCellValue(idOrName: string): any
    getCellValueAsString(idOrName: string): string
}

export interface RecordQueryResult {
    recordIds: string[],
    records: Record[],
}
