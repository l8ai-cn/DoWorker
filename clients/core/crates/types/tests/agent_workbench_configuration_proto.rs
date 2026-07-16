use agentsmesh_types::proto_agent_workbench_v2 as v2;

#[test]
fn configuration_contract_is_generated_across_snapshot_events_and_runner_ingress() {
    let configuration = v2::SessionConfiguration {
        model: Some("gpt-5.5".into()),
        permission_mode: Some("default".into()),
    };
    let snapshot = v2::SessionSnapshot {
        configuration: Some(configuration.clone()),
        ..Default::default()
    };
    let event = v2::agent_event::Event::ConfigurationChanged(v2::ConfigurationChanged {
        configuration: Some(configuration.clone()),
    });
    let mutation = v2::runner_workbench_mutation::Mutation::Configuration(configuration.clone());

    assert_eq!(snapshot.configuration, Some(configuration.clone()));
    assert!(matches!(
        event,
        v2::agent_event::Event::ConfigurationChanged(_)
    ));
    assert!(matches!(
        mutation,
        v2::runner_workbench_mutation::Mutation::Configuration(_)
    ));
}
