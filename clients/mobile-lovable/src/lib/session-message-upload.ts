import { uploadSessionFile, type UploadedFile } from "./files-api";
import { postMessageContent, type MessageContentBlock } from "./sessions-api";

export interface SessionMessageAttachment {
  file: File;
  uploaded?: UploadedFile;
}

export async function postSessionMessageWithFiles(
  sessionId: string,
  text: string,
  attachments: SessionMessageAttachment[],
): Promise<void> {
  const trimmedText = text.trim();
  if (!trimmedText && attachments.length === 0) {
    throw new Error("请输入消息或添加附件");
  }
  const uploads = await Promise.all(
    attachments.map(async (attachment) => {
      if (attachment.uploaded) return attachment.uploaded;
      const upload = await uploadSessionFile(sessionId, attachment.file);
      attachment.uploaded = upload;
      return upload;
    }),
  );
  const content: MessageContentBlock[] = uploads.map((upload, index) => ({
    type: attachments[index].file.type.startsWith("image/") ? "input_image" : "input_file",
    file_id: upload.id,
    filename: upload.filename,
  }));
  if (trimmedText) {
    content.push({ type: "input_text", text: trimmedText });
  }
  await postMessageContent(sessionId, content);
}
