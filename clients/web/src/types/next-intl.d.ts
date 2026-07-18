import common from "@/messages/en/common.json";
import auth from "@/messages/en/auth.json";
import landing from "@/messages/en/landing.json";
import workforce from "@/messages/en/workforce.json";
import app from "@/messages/en/app.json";
import settings from "@/messages/en/settings.json";
import ide from "@/messages/en/ide.json";
import repositories from "@/messages/en/repositories.json";
import runners from "@/messages/en/runners.json";
import docs from "@/messages/en/docs.json";
import content from "@/messages/en/content.json";
import extensions from "@/messages/en/extensions.json";
import workflows from "@/messages/en/workflows.json";
import channels from "@/messages/en/channels.json";
import blockstore from "@/messages/en/blockstore.json";
import infra from "@/messages/en/infra.json";
import automation from "@/messages/en/automation.json";
import experts from "@/messages/en/experts.json";
import resourceOrchestration from "@/messages/en/resource-orchestration.json";
import changelogEntries from "@/messages/en/changelog-entries.json";
import videoWorker from "@/messages/en/video-worker.json";

type Messages = typeof common &
  typeof auth &
  typeof landing &
  typeof workforce &
  typeof app &
  typeof settings &
  typeof ide &
  typeof repositories &
  typeof runners &
  typeof docs &
  typeof content &
  typeof extensions &
  typeof workflows &
  typeof channels &
  typeof blockstore &
  typeof infra &
  typeof automation &
  typeof experts &
  typeof resourceOrchestration &
  typeof changelogEntries &
  typeof videoWorker;

declare global {
  // eslint-disable-next-line @typescript-eslint/no-empty-object-type
  interface IntlMessages extends Messages {}
}
