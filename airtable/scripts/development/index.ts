import { pushPayloads } from "./readWriteTransactions";

function main() {
    pushPayloads(5).then(() => console.log("Done."));
}

main();
