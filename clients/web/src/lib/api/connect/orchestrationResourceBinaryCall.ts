import { fromBinary, type Message } from "@bufbuild/protobuf";
import type { GenMessage } from "@bufbuild/protobuf/codegenv2";

type BinaryMethod = (request: Uint8Array) => Promise<Uint8Array>;

export async function callOrchestrationResource<T extends Message>(
  responseSchema: GenMessage<T>,
  request: Uint8Array,
  method: BinaryMethod,
): Promise<T> {
  const response = await method(request);
  return fromBinary(responseSchema, new Uint8Array(response));
}
