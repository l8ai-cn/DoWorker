use crate::core::AgentsMeshCore;
use crate::dto::{
    AuthorizeRunnerRequestDto, CreateRunnerTokenRequestDto, GrpcRegistrationTokenDto,
    PodListResponseDto, RunnerAuthStatusDto, RunnerDto, RunnerListResponseDto,
    RunnerLogListResponseDto, RunnerTokenListResponseDto, UpdateRunnerRequestDto,
    UpgradeRunnerRequestDto,
};
use crate::error::CoreError;

#[uniffi::export(async_runtime = "tokio")]
impl AgentsMeshCore {
    pub async fn list_runners(
        &self,
        status: Option<String>,
    ) -> Result<RunnerListResponseDto, CoreError> {
        let resp = self.api.list_runners(status.as_deref()).await?;
        Ok(resp.into())
    }

    pub async fn list_available_runners(&self) -> Result<RunnerListResponseDto, CoreError> {
        let resp = self.api.list_available_runners().await?;
        Ok(resp.into())
    }

    pub async fn get_runner(&self, id: i64) -> Result<RunnerDto, CoreError> {
        let resp = self.api.get_runner(id).await?;
        Ok(resp.runner.into())
    }

    pub async fn update_runner(
        &self,
        id: i64,
        req: UpdateRunnerRequestDto,
    ) -> Result<RunnerDto, CoreError> {
        let runner = self.api.update_runner(id, &req.into()).await?;
        Ok(runner.into())
    }

    pub async fn delete_runner(&self, id: i64) -> Result<(), CoreError> {
        self.api.delete_runner(id).await?;
        Ok(())
    }

    pub async fn create_runner_token(
        &self,
        req: CreateRunnerTokenRequestDto,
    ) -> Result<GrpcRegistrationTokenDto, CoreError> {
        let tok = self.api.create_runner_token(&req.into()).await?;
        Ok(tok.into())
    }

    pub async fn list_runner_tokens(&self) -> Result<RunnerTokenListResponseDto, CoreError> {
        let resp = self.api.list_runner_tokens().await?;
        Ok(resp.into())
    }

    pub async fn delete_runner_token(&self, id: i64) -> Result<(), CoreError> {
        self.api.delete_runner_token(id).await?;
        Ok(())
    }

    pub async fn list_runner_pods(
        &self,
        id: i64,
        status: Option<String>,
        limit: Option<u32>,
        offset: Option<u32>,
    ) -> Result<PodListResponseDto, CoreError> {
        let resp = self
            .api
            .list_runner_pods(id, status.as_deref(), limit, offset)
            .await?;
        Ok(resp.into())
    }

    pub async fn upgrade_runner(
        &self,
        id: i64,
        req: UpgradeRunnerRequestDto,
    ) -> Result<(), CoreError> {
        self.api.upgrade_runner(id, &req.into()).await?;
        Ok(())
    }

    pub async fn request_runner_log_upload(&self, id: i64) -> Result<(), CoreError> {
        self.api.request_runner_log_upload(id).await?;
        Ok(())
    }

    pub async fn list_runner_logs(&self, id: i64) -> Result<RunnerLogListResponseDto, CoreError> {
        let resp = self.api.list_runner_logs(id).await?;
        Ok(resp.into())
    }

    pub async fn get_runner_auth_status(
        &self,
        auth_key: String,
    ) -> Result<RunnerAuthStatusDto, CoreError> {
        let status = self.api.get_runner_auth_status(&auth_key).await?;
        Ok(status.into())
    }

    pub async fn authorize_runner(
        &self,
        req: AuthorizeRunnerRequestDto,
    ) -> Result<(), CoreError> {
        self.api.authorize_runner(&req.into()).await?;
        Ok(())
    }
}
