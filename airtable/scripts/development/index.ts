import { mergeTransactions } from "./compareTransactions";

function main() {
    mergeTransactions().then(({ atu, gbu }) => {
        console.log(`Unmatched from Airtable:\n${JSON.stringify(atu, null, 2)}`);
        console.log(`Unmatched from Givebutter:\n${JSON.stringify(gbu, null, 2)}`);
    });
}

main();
