/**
 * Neutralises spreadsheet formula injection (OWASP "CSV Injection") in an exported cell.
 *
 * Participant free text (daily `feelings`, about-me answers, usernames) can start with
 * `=`, `+`, `-`, `@`, tab or carriage return; Excel/Sheets would evaluate such a cell as a
 * formula when a researcher opens the export (e.g. =HYPERLINK/WEBSERVICE exfiltration).
 * Prefixing a `'` makes the spreadsheet treat it as literal text. Only string values are
 * touched — numbers (including negatives), booleans and UUIDs export unchanged.
 */
export function neutralizeCsvFormula(cell: string, isText: boolean): string {
    return isText && /^[=+\-@\t\r]/.test(cell) ? `'${cell}` : cell;
}
