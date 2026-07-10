import { create, fromBinary, toBinary, type Message } from "@bufbuild/protobuf";
import type { GenMessage } from "@bufbuild/protobuf/codegenv2";
import {
  CreateOrganizationConnectionRequestSchema, CreatePersonalConnectionRequestSchema,
  CreateResourceRequestSchema, DeleteConnectionRequestSchema, DeleteResourceRequestSchema,
  GetCatalogRequestSchema, GetCatalogResponseSchema, ListConnectionsResponseSchema,
  ListOrganizationConnectionsRequestSchema, ListOrganizationEffectiveResourcesRequestSchema,
  ListPersonalConnectionsRequestSchema, ListPersonalEffectiveResourcesRequestSchema,
  ListEffectiveResourcesResponseSchema, MutationResponseSchema,
  RotateConnectionCredentialsRequestSchema, SetConnectionEnabledRequestSchema,
  SetDefaultRequestSchema, SetResourceEnabledRequestSchema, UpdateConnectionRequestSchema,
  UpdateResourceRequestSchema, ValidateConnectionRequestSchema,
} from "@proto/ai_resource/v1/ai_resource_pb";
import {
  ConnectionUpdateSchema, ModelResourceSchema, ProviderConnectionSchema,
  ResourceSpecSchema, ResourceUpdateSchema,
} from "@proto/ai_resource/v1/types_pb";
import { getAIResourceService } from "@/lib/wasm-core";
import {
  fromEffectiveResource, fromModelResource, fromProviderConnection, fromProviderDefinition,
  type ConnectionInput, type ResourceInput,
} from "../facade/aiResource";

type BinaryMethod = (request: Uint8Array) => Promise<Uint8Array>;

async function call<T extends Message>(
  responseSchema: GenMessage<T>, request: Uint8Array, method: BinaryMethod,
): Promise<T> {
  const response = await method(request);
  return fromBinary(responseSchema, new Uint8Array(response));
}

export async function getCatalog() {
  const request = toBinary(GetCatalogRequestSchema, create(GetCatalogRequestSchema));
  const response = await call(GetCatalogResponseSchema, request, (bytes) => getAIResourceService().getCatalogConnect(bytes));
  return response.providers.map(fromProviderDefinition);
}

export async function listPersonalConnections() {
  const request = toBinary(ListPersonalConnectionsRequestSchema, create(ListPersonalConnectionsRequestSchema));
  const response = await call(ListConnectionsResponseSchema, request, (bytes) => getAIResourceService().listPersonalConnectionsConnect(bytes));
  return response.connections.map(fromProviderConnection);
}

export async function listOrganizationConnections(orgSlug: string) {
  const request = toBinary(ListOrganizationConnectionsRequestSchema, create(ListOrganizationConnectionsRequestSchema, { orgSlug }));
  const response = await call(ListConnectionsResponseSchema, request, (bytes) => getAIResourceService().listOrganizationConnectionsConnect(bytes));
  return response.connections.map(fromProviderConnection);
}

export async function listPersonalEffectiveResources(modalities: string[] = []) {
  const request = toBinary(ListPersonalEffectiveResourcesRequestSchema, create(ListPersonalEffectiveResourcesRequestSchema, { modalities }));
  const response = await call(ListEffectiveResourcesResponseSchema, request, (bytes) => getAIResourceService().listPersonalEffectiveResourcesConnect(bytes));
  return response.resources.map(fromEffectiveResource);
}

export async function listOrganizationEffectiveResources(orgSlug: string, modalities: string[] = []) {
  const request = toBinary(ListOrganizationEffectiveResourcesRequestSchema, create(ListOrganizationEffectiveResourcesRequestSchema, { orgSlug, modalities }));
  const response = await call(ListEffectiveResourcesResponseSchema, request, (bytes) => getAIResourceService().listOrganizationEffectiveResourcesConnect(bytes));
  return response.resources.map(fromEffectiveResource);
}

export async function createPersonalConnection(input: ConnectionInput) {
  const request = toBinary(CreatePersonalConnectionRequestSchema, create(CreatePersonalConnectionRequestSchema, input));
  return fromProviderConnection(await call(ProviderConnectionSchema, request, (bytes) => getAIResourceService().createPersonalConnectionConnect(bytes)));
}

export async function createOrganizationConnection(input: ConnectionInput & { orgSlug: string }) {
  const request = toBinary(CreateOrganizationConnectionRequestSchema, create(CreateOrganizationConnectionRequestSchema, input));
  return fromProviderConnection(await call(ProviderConnectionSchema, request, (bytes) => getAIResourceService().createOrganizationConnectionConnect(bytes)));
}

export async function updateConnection(connectionId: number, input: { name: string; baseUrl: string; credentials?: Record<string, string> }) {
  const connection = create(ConnectionUpdateSchema, { ...input, hasCredentials: input.credentials !== undefined, credentials: input.credentials ?? {} });
  const request = toBinary(UpdateConnectionRequestSchema, create(UpdateConnectionRequestSchema, { connectionId: BigInt(connectionId), connection }));
  return fromProviderConnection(await call(ProviderConnectionSchema, request, (bytes) => getAIResourceService().updateConnectionConnect(bytes)));
}

async function mutate<T extends Message>(schema: GenMessage<T>, value: T, method: BinaryMethod) {
  await call(MutationResponseSchema, toBinary(schema, value), method);
}

export const rotateConnectionCredentials = (connectionId: number, credentials: Record<string, string>) => mutate(RotateConnectionCredentialsRequestSchema, create(RotateConnectionCredentialsRequestSchema, { connectionId: BigInt(connectionId), credentials }), (bytes) => getAIResourceService().rotateConnectionCredentialsConnect(bytes));
export const setConnectionEnabled = (connectionId: number, enabled: boolean) => mutate(SetConnectionEnabledRequestSchema, create(SetConnectionEnabledRequestSchema, { connectionId: BigInt(connectionId), enabled }), (bytes) => getAIResourceService().setConnectionEnabledConnect(bytes));
export const validateConnection = (connectionId: number) => mutate(ValidateConnectionRequestSchema, create(ValidateConnectionRequestSchema, { connectionId: BigInt(connectionId) }), (bytes) => getAIResourceService().validateConnectionConnect(bytes));
export const deleteConnection = (connectionId: number) => mutate(DeleteConnectionRequestSchema, create(DeleteConnectionRequestSchema, { connectionId: BigInt(connectionId) }), (bytes) => getAIResourceService().deleteConnectionConnect(bytes));

export async function createResource(connectionId: number, input: ResourceInput) {
  const resource = create(ResourceSpecSchema, input);
  const request = toBinary(CreateResourceRequestSchema, create(CreateResourceRequestSchema, { connectionId: BigInt(connectionId), resource }));
  return fromModelResource(await call(ModelResourceSchema, request, (bytes) => getAIResourceService().createResourceConnect(bytes)));
}

export async function updateResource(resourceId: number, input: Omit<ResourceInput, "identifier">) {
  const resource = create(ResourceUpdateSchema, input);
  const request = toBinary(UpdateResourceRequestSchema, create(UpdateResourceRequestSchema, { resourceId: BigInt(resourceId), resource }));
  return fromModelResource(await call(ModelResourceSchema, request, (bytes) => getAIResourceService().updateResourceConnect(bytes)));
}

export const setResourceEnabled = (resourceId: number, enabled: boolean) => mutate(SetResourceEnabledRequestSchema, create(SetResourceEnabledRequestSchema, { resourceId: BigInt(resourceId), enabled }), (bytes) => getAIResourceService().setResourceEnabledConnect(bytes));
export const deleteResource = (resourceId: number) => mutate(DeleteResourceRequestSchema, create(DeleteResourceRequestSchema, { resourceId: BigInt(resourceId) }), (bytes) => getAIResourceService().deleteResourceConnect(bytes));
export const setDefaultResource = (resourceId: number, modality: string) => mutate(SetDefaultRequestSchema, create(SetDefaultRequestSchema, { resourceId: BigInt(resourceId), modality }), (bytes) => getAIResourceService().setDefaultConnect(bytes));
