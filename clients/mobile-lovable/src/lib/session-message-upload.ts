import { uploadSessionFile } from "./files-api";
import { postMessageContent, type MessageContentBlock } from "./sessions-api";

export async function postSessionMessageWithFiles(
  sessionId: string,
  text: string,
  files: File[],
): Promise<void> {
  const trimmedText = text.trim();
  if (!trimmedText && files.length === 0) {
    throw new Error("请输入消息或添加附件");
  }
  const uploads = await Promise.all(files.map((file) => uploadSessionFile(sessionId, file)));
  const content: MessageContentBlock[] = uploads.map((upload, index) => ({
    type: files[index].type.startsWith("image/") ? "input_image" : "input_file",
    file_id: upload.id,
    filename: upload.filename,
  }));
  if (trimmedText) {
    content.push({ type: "input_text", text: trimmedText });
  }
  await postMessageContent(sessionId, content);
}
