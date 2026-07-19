import { NewChatLandingView } from "./NewChatLandingView";
import { useNewChatLandingController } from "./useNewChatLandingController";

export function NewChatLandingScreen() {
  return <NewChatLandingView state={useNewChatLandingController()} />;
}
