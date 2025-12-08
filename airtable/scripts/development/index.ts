import { loadPayloads } from "./readWriteTransactions";
import { ListDonationsBetweenDates } from "./analyzeGbData";

function main() {
    const payloads = loadPayloads();
    const ids = ListDonationsBetweenDates(
        payloads,
        new Date("2025-11-24T00:00:00.000Z"),
        new Date("2025-11-25T00:00:00.000Z")
    );
    console.log(`There are ${ids.length} payloads:`);
    console.log(ids);
}

main();
