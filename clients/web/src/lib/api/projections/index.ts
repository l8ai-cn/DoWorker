// Aggregated proto→viewModel projections used by Web Connect adapters and
// state stores.
export { loopToCache, loopRunToCache } from "./loop";
export {
  ticketToCache,
  boardColumnToCache,
  labelToCache,
  cacheTicketToProto,
} from "./ticket";
export { podToCache } from "./pod";
export { runnerToCache } from "./runner";
export { repositoryToCache, branchToCache } from "./repository";
