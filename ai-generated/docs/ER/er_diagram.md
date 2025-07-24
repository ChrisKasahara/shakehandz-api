```mermaid
erDiagram
    proposals ||--|| projects : "project_id"
    proposals ||--|| human_resources : "human_resource_id"
    proposals ||--|| flow_master : "flow_id"
    proposals ||--|| result_master : "result_id"
    proposals ||--|| result_detail_master : "result_detail_id"
    proposals ||--|| users : "proposer_id"
    projects ||--|| clients : "client_id"
    human_resources ||--|| clients : "client_id"
    result_detail_master ||--|| result_master : "result_id"

    proposals {
        id string PK
        proposal_date datetime
        proposer_id string FK
        project_id string FK
        human_resource_id string FK
        flow_id int FK
        result_id int FK
        result_detail_id int FK
        memo string
    }

    projects {
        id string PK
        email_id string
        email_subject string
        email_sender string
        email_received_at datetime
        project_start_month datetime
        prefecture string
        work_location string
        remote_work_frequency string
        working_hours string
        required_skills string
        unit_price_min int
        unit_price_max int
        unit_price_unit string
        business_flow stringp
        business_flow_restrictions string
        priority_talent string
        project_summary string
        registered_at datetime
        extraction_confidence float
        extraction_notes string
        client_id string FK
    }

    human_resources {
        id string PK
        email_id string
        email_subject string
        email_sender string
        email_received_at datetime
        attachment_filename string
        candidate_initial string
        age int
        prefecture string
        nearest_station string
        work_conditions string
        employment_type string
        main_lang_fw string
        main_skills string
        experience_phases string
        available_start_month string
        hourly_rate_min int
        hourly_rate_max int
        hourly_rate_unit string
        additional_info string
        extraction_confidence float
        extraction_notes string
        career_summary string
        registered_at datetime
        client_id string FK
    }

    clients {
        id string PK
        name string
    }

    flow_master {
        id int PK
        name string
    }

    result_master {
        id int PK
        name string
    }

    result_detail_master {
        id int PK
        result_id int FK
        description string
    }

    users {
        id string PK
        google_id string UNIQUE NOT NULL
        email string NOT NULL
        name string
        picture string
    }
```
