# ER 図

```mermaid
erDiagram

    HUMAN_RESOURCE {
        uint ID PK
        string MessageID
        string? AttachmentType
        string? AttachmentFilename
        string EmailReceivedAt
        string? ProviderCompany
        string? SalesPerson
        string? CandidateInitial
        uint8? Age
        enum? Nationality
        json Roles
        json ExperienceAreas
        json MainSkills
        json SubSkills
        string? AdditionalInfo
        enum? EmploymentType
        enum? WorkStyle
        bool IsDirectlyUnder
    }

    PROJECT {
        string ID PK
        string EmailID
        string? EmailSubject
        string? EmailSender
        time? EmailReceivedAt
        time? ProjectStartMonth
        string? Prefecture
        string? WorkLocation
        string? RemoteWorkFrequency
        string? WorkingHours
        string? RequiredSkills
        uint? UnitPriceMin
        uint? UnitPriceMax
        string? UnitPriceUnit
        string? BusinessFlow
        string? BusinessFlowRestrictions
        string? PriorityTalent
        string? ProjectSummary
        time? RegisteredAt
        float64? ExtractionConfidence
        string? ExtractionNotes
    }

    %% 関連付け（現状、直接的なリレーションはなし）
```
