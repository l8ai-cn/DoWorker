import { invoke } from "./invoke";
import type { IMessageService } from "@agentsmesh/service-interface";

export class ElectronMessageService implements IMessageService {
  async get_messages(unreadOnly?: boolean | null, limit?: number | null, offset?: number | null): Promise<string> {
    return invoke<string>("messageGetMessages", unreadOnly, limit, offset);
  }

  async get_message(id: bigint): Promise<string> {
    return invoke<string>("messageGetMessage", Number(id));
  }

  async get_sent_messages(limit?: number | null, offset?: number | null): Promise<string> {
    return invoke<string>("messageGetSentMessages", limit, offset);
  }

  async get_conversation(correlationId: string, limit?: number | null): Promise<string> {
    return invoke<string>("messageGetConversation", correlationId, limit);
  }

  async get_unread_count(): Promise<string> {
    return invoke<string>("messageGetUnreadCount");
  }

  async send_message(json: string, podKey?: string | null): Promise<string> {
    return invoke<string>("messageSendMessage", json, podKey);
  }

  async mark_read(json: string): Promise<string> {
    return invoke<string>("messageMarkRead", json);
  }

  async mark_all_read(): Promise<string> {
    return invoke<string>("messageMarkAllRead");
  }

  async get_dead_letters(limit?: number | null, offset?: number | null): Promise<string> {
    return invoke<string>("messageGetDeadLetters", limit, offset);
  }

  async replay_dead_letter(entryId: bigint): Promise<string> {
    return invoke<string>("messageReplayDeadLetter", Number(entryId));
  }
}
